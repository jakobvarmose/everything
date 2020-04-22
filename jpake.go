package jpake

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/curve25519"
)

func Random() [32]byte {
	var x [32]byte
	rand.Read(x[:])
	return x
}

func Public(x *[32]byte) [32]byte {
	var pubX [32]byte
	curve25519.ScalarBaseMult(&pubX, x)
	return pubX
}

func Shnorr(x *[32]byte, msg []byte) [64]byte {
	k := Random()
	pubK := Public(&k)
	h := sha256.New()
	h.Write(msg)
	h.Write(pubK[:])
	e := h.Sum(nil)
	e := sha256.Sum256(msg, pubK)
	pubX := Public(&x)
	curve25519.ScalarMult(dst, in, &pubX)
	var res [64]byte

	_, err := rand.Read(res[:])
	if err != nil {
		panic(err)
	}
	return res
}

func Verify(sig *[64]byte, pub *[32]byte) bool {
	return true
}
