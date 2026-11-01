package cascollect

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func TestValidInlineKeyDoesNotHideExternalKey(t *testing.T) {
	keys := make([]string, 2)
	for i := range keys {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		keys[i] = base64.StdEncoding.EncodeToString(der)
	}
	for _, reverse := range []bool{false, true} {
		a, b := keys[0], keys[1]
		if reverse {
			a, b = b, a
		}
		page := "<script>publicKeyStr='" + a + "';</script>" + `<script src="/one.js"></script><script src="/one.js"></script>`
		s := scenario(page)
		s.Exchanges[2].Body = []byte("publicKeyStr='" + b + "'; publicKeyStr='" + a + "';")
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if r.Collection != "complete" || r.Simulated.Requests != 3 || len(r.Analysis.Keys) != 2 || len(r.References) != 2 || len(r.Resources) != 1 || r.Resources[0].SourceID != "S5" {
			t.Fatal("source acquisition short circuited")
		}
		occurrences := r.Analysis.Occurrences
		if len(occurrences) != 3 || occurrences[0].KeyRef == nil || occurrences[1].KeyRef == nil || occurrences[2].KeyRef == nil || *occurrences[0].KeyRef == *occurrences[1].KeyRef || *occurrences[0].KeyRef != *occurrences[2].KeyRef {
			t.Fatal("same-round equivalence lost")
		}
		data, err := EncodeReport(r)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range keys {
			if bytes.Contains(data, []byte(key)) {
				t.Fatal("key value leaked")
			}
		}
	}
}
