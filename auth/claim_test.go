package auth

import (
	"encoding/json"
	"testing"
)

func TestClaimNumber(t *testing.T) {
	n, ok := claimNumber(float64(12))
	if !ok || n != 12 {
		t.Fatalf("float64 = %v ok = %v", n, ok)
	}
	n, ok = claimNumber(json.Number("34"))
	if !ok || n != 34 {
		t.Fatalf("json.Number = %v ok = %v", n, ok)
	}
	if _, ok = claimNumber("nope"); ok {
		t.Fatal("string was accepted")
	}
	if _, ok = claimNumber(nil); ok {
		t.Fatal("nil was accepted")
	}
	if _, ok = claimNumber(json.Number("nope")); ok {
		t.Fatal("bad json.Number was accepted")
	}
}
