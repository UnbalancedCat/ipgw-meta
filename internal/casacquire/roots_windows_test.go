//go:build windows

package casacquire

import (
	"reflect"
	"testing"
)

// Only reads public local certificates; no profiles, private keys or network.
func TestLocalRootPoolCannotUseSystemChainRetrieval(t *testing.T) {
	snapshot := &rootSnapshot{}
	pool, denied, ok := localRoots(snapshot)
	if !ok || pool == nil || denied == nil {
		t.Fatal("local certificate stores unavailable")
	}
	// The pinned Go 1.25.0 implementation selects CryptoAPI by this flag.
	// Guard the no-hidden-network property when changing root loading.
	flag := reflect.ValueOf(pool).Elem().FieldByName("systemPool")
	if !flag.IsValid() || flag.Bool() {
		t.Fatal("system chain verifier enabled")
	}
	if snapshot.failure != "" {
		if snapshot.certs != nil || (snapshot.failure != "root_snapshot_limit" && snapshot.failure != "root_snapshot_invalid") {
			t.Fatal("invalid optional root snapshot failure")
		}
	} else {
		reconstructed, _ := relationRoots(snapshot.certs...)
		if !pool.Equal(reconstructed) {
			t.Fatal("diagnostic root snapshot diverged from original pool")
		}
	}
}
