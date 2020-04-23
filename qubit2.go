package main

import (
	"fmt"
	"math/rand"
)

type Entanglement struct {
	qubits []*Qubit
	data   []complex64
}

func combine(a, b *Entanglement) {
	if a == b {
		return
	}
	c := &Entanglement{
		qubits: make([]*Qubit, len(a.qubits)+len(b.qubits)),
		data:   make([]complex64, len(a.data)*len(b.data)),
	}
	for i := range a.qubits {
		c.qubits[i] = a.qubits[i]
		c.qubits[i].ent = c
	}
	for i := range b.qubits {
		c.qubits[len(a.qubits)+i] = b.qubits[i]
		c.qubits[len(a.qubits)+i].ent = c
		c.qubits[len(a.qubits)+i].shift = uint(len(a.qubits) + i)
	}
	for j := range b.data {
		for i := range a.data {
			c.data[j*len(a.data)+i] = a.data[i] * b.data[j]
		}
	}
}

func extract(e *Entanglement, shift uint) (zero, one complex64) {
	for i := 0; i < len(e.data); i++ {
		if (i>>shift)&1 == 0 {
			zero += e.data[i]
		} else {
			one += e.data[i]
		}
	}
	fmt.Println(zero, one)
	return
}

type Qubit struct {
	ent   *Entanglement
	shift uint
}

func New(zero, one complex64) *Qubit {
	q := &Qubit{}
	e := &Entanglement{
		qubits: []*Qubit{q},
		data:   []complex64{zero, one},
	}
	q.ent = e
	return q
}

func (q *Qubit) CNot(t *Qubit) {
	combine(q.ent, t.ent)

}

func (q *Qubit) Measure() int {
	r := rand.Float32()
	extract(q.ent, q.shift)
	fmt.Println(r)
	return 0
}
