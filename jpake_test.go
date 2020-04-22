package jpake

import "testing"

func TestRandom(t *testing.T) {
	if Random() == Random() {
		t.Error("Two random keys are equal")
	}
}

func TestPublic(t *testing.T) {
	x := Random()
	if Public(&x) == x {
		t.Error("Public key equal to private key")
	}
	if Public(&x) != Public(&x) {
		t.Error("Identical public keys are not equal")
	}
	y := Random()
	if Public(&x) == Public(&y) {
		t.Error("Different public keys are equal")
	}
}

func TestShnorr(t *testing.T) {
	x := Random()
	msg := []byte{1}
	if Shnorr(&x, msg) == Shnorr(&x, msg) {
		t.Error("Two signatures are equal")
	}
}

func TestVerify(t *testing.T) {
	priv := Random()
	msg := []byte{1}
	sig := Shnorr(&priv, msg)
	pub := Public(&priv)
	if !Verify(&sig, &pub) {
		t.Error("Verification failed")
	}
	priv2 := Random()
	pub2 := Public(&priv2)
	if Verify(&sig, &pub2) {
		t.Error("Verification succeded")
	}
}
