package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

// merkle signatures

type key struct {
	x [512][32]byte
	p pubkey
}

type pubkey [32]byte

func hash(x [][32]byte) [32]byte {
	if len(x) == 1 {
		return sha256.Sum256(x[0][:])
	}
	a := hash(x[:len(x)/2])
	b := hash(x[len(x)/2:])
	return sha256.Sum256(append(a[:], b[:]...))
}

func GenerateKey() key {
	var x [512][32]byte
	for i := range x {
		_, err := rand.Read(x[i][:])
		if err != nil {
			panic(err)
		}
	}
	p := hash(x[:]) // This is not stadard - we shouldn't build a tree
	return key{x, p}
}

type signature struct {
	h [32]byte
	s [256][32]byte
	y [512][32]byte
}

func (k *key) ToPublic() pubkey {
	return k.p
}

func getBit(x [32]byte, i int) int {
	return int(x[i/8]>>uint(i%8)) & 1
}

func (k *key) Sign(msg []byte) signature {
	h := sha256.Sum256(msg)
	var s [256][32]byte
	var y [512][32]byte
	for i := range s {
		b := getBit(h, i)
		s[i] = k.x[2*i+b]
	}
	for i := range y {
		y[i] = sha256.Sum256(k.x[i][:])
	}
	return signature{h, s, y}
}

func main() {
	k := GenerateKey()
	p := k.ToPublic()
	fmt.Printf("%x\n", p)
}
