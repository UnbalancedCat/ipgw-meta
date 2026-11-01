package casacquire

import (
	"net/netip"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func TestDNSFuzzSVCBPanicRejected(t *testing.T) {
	// Minimized, generated fuzz input a92e9bef76c5ada7; no captured traffic.
	packet := []byte("0000\x00\x00000000\x00\x00A0000000000\x0000\x00\x00000%")
	name, _ := dnsmessage.NewName("example.invalid.")
	if _, ok := dnsAnswer(packet, 7, name); ok {
		t.Fatal("malformed fuzz input accepted")
	}
}

func TestDNSOnlyInterpretsAddressAnswerSubset(t *testing.T) {
	name, _ := dnsmessage.NewName("example.invalid.")
	for _, section := range []string{"answer", "authority", "additional"} {
		for _, kind := range []dnsmessage.Type{64, 65} {
			m := dnsmessage.Message{Header: dnsmessage.Header{ID: 7, Response: true}, Questions: []dnsmessage.Question{{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}, Answers: []dnsmessage.Resource{{Header: dnsmessage.ResourceHeader{Name: name, Class: dnsmessage.ClassINET}, Body: &dnsmessage.AResource{A: [4]byte{192, 0, 2, 1}}}}}
			opaque := dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: name, Class: dnsmessage.ClassINET}, Body: &dnsmessage.UnknownResource{Type: kind, Data: []byte{0, 0, 0, 0, 0, 0x30, 0x30}}}
			switch section {
			case "answer":
				m.Answers = append(m.Answers, opaque)
			case "authority":
				m.Authorities = []dnsmessage.Resource{opaque}
			case "additional":
				m.Additionals = []dnsmessage.Resource{opaque}
			}
			packet, err := m.Pack()
			if err != nil {
				t.Fatal(err)
			}
			if got, ok := dnsAnswer(packet, 7, name); !ok || got != netip.MustParseAddr("192.0.2.1") {
				t.Fatal("unrelated RDATA affected A result")
			}
			if _, ok := dnsAnswer(packet[:len(packet)-1], 7, name); ok {
				t.Fatal("truncated opaque record accepted")
			}
		}
	}
	for _, data := range [][]byte{{192, 0, 2}, {192, 0, 2, 1, 0}} {
		m := dnsmessage.Message{Header: dnsmessage.Header{ID: 7, Response: true}, Questions: []dnsmessage.Question{{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}, Answers: []dnsmessage.Resource{{Header: dnsmessage.ResourceHeader{Name: name, Class: dnsmessage.ClassINET}, Body: &dnsmessage.UnknownResource{Type: dnsmessage.TypeA, Data: data}}}}
		packet, err := m.Pack()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := dnsAnswer(packet, 7, name); ok {
			t.Fatal("invalid A record length accepted")
		}
	}
}
