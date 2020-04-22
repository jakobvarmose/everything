package main

import (
	"errors"
	"math/rand"
)

var (
	EntanglementLimit int = 27
)

type entanglement struct {
	states []complex64
	qubits []*qubit
}

func (e *entanglement) join(o *entanglement) {
	if e == o {
		return
	}
	if len(e.qubits)+len(o.qubits) > EntanglementLimit {
		panic(errors.New("entanglement limit exceeded"))
	}
	states := make([]complex64, len(e.states)*len(o.states))
	for i := range states {
		j := i % len(e.states)
		k := i / len(e.states)
		states[i] = e.states[j] * o.states[k]
	}
	e.qubits = append(e.qubits, o.qubits...)
	for i, qubit := range e.qubits {
		qubit.e = e
		qubit.i = i
	}
	e.states = states
}

func (e *entanglement) remove(i int, aa, bb complex64) {
	s := 1 << uint(i)
	m := len(e.qubits) - 1
	t := 1 << uint(m)
	states := make([]complex64, len(e.states)/2)
	for j := range states {
		if j&s == 0 {
			//FIXME can they be added here???
			if aa != 0 {
				states[j] = e.states[j]
			} else {
				states[j] = e.states[j|s]
			}
		} else {
			if aa != 0 {
				states[j] = e.states[j&^s|t]
			} else {
				states[j] = e.states[j|t]
			}
		}
	}
	e.states = states
	if i != m {
		e.qubits[i] = e.qubits[m]
		e.qubits[i].i = i
	}
	e.qubits = e.qubits[:m]
}

func (e *entanglement) apply(i int, gate Gate) {
	Clock++
	s := 1 << uint(i)
	var failed bool
	ratio := float32(-1.0)
	var aa, bb complex64
	for j := range e.states {
		if j&s == 0 {
			a := e.states[j]
			b := e.states[j|s]
			e.states[j] = gate[0]*a + gate[2]*b
			e.states[j|s] = gate[1]*a + gate[3]*b
			zero := real(e.states[j] * complex(real(e.states[j]), -imag(e.states[j])))
			one := real(e.states[j|s] * complex(real(e.states[j|s]), -imag(e.states[j|s])))
			if e.states[j] != 0 || e.states[j|s] != 0 {
				if ratio == -1 {
					ratio = one / (zero + one)
					aa = e.states[j]
					bb = e.states[j|s]
				}
				if e.states[j] != aa || e.states[j|s] != bb {
					failed = true
				}
			}
		}
	}
	if len(e.qubits) != 1 && !failed {
		//println("RATIO", aa, bb)
		aa, bb = aa, bb
		//fmt.Println("OPTIMIZE", len(e.qubits), i)
		/*q := e.qubits[i]
		q.e = &entanglement{
			states: []complex64{aa, bb}, //FIXME normalize
			qubits: []*qubit{q},
		}
		q.i = 0*/
		//e.remove(i, aa, bb) //FIXME normalize
	}
}

func (e *entanglement) control(i, j int, gate Gate) {
	Clock++
	si := 1 << uint(i)
	sj := 1 << uint(j)
	for k := range e.states {
		if k&si == 0 && k&sj == 0 {
			//TODO add optimization to split entanglement
			a := e.states[k|si]
			b := e.states[k|si|sj]
			e.states[k|si] = gate[0]*a + gate[2]*b
			e.states[k|si|sj] = gate[1]*a + gate[3]*b
		}
	}
}

func (e *entanglement) measure(i int) int {
	Clock++
	var zero, one complex64
	s := 1 << uint(i)
	for j := range e.states {
		if j&s == 0 {

			zero += e.states[j] * complex(real(e.states[j]), -imag(e.states[j]))
			one += e.states[j|s] * complex(real(e.states[j|s]), -imag(e.states[j|s]))
		}
	}
	f := rand.Float32() * (real(zero) + real(one))
	if f < real(zero) {
		m := len(e.qubits) - 1
		q := e.qubits[i]
		q.e = &entanglement{
			states: []complex64{1, 0},
			qubits: []*qubit{q},
		}
		q.i = 0

		states := make([]complex64, len(e.states)/2)
		t := 1 << uint(m)
		for j := range states {
			if j&s == 0 {
				states[j] = e.states[j]
			} else {
				states[j] = e.states[j&^s|t]
			}
		}
		e.states = states
		if i != m {
			e.qubits[i] = e.qubits[m]
			e.qubits[i].i = i
		}
		e.qubits = e.qubits[:m]
		return 0
	} else {
		m := len(e.qubits) - 1
		q := e.qubits[i]
		q.e = &entanglement{
			states: []complex64{0, 1},
			qubits: []*qubit{q},
		}
		q.i = 0

		states := make([]complex64, len(e.states)/2)
		t := 1 << uint(m)
		for j := range states {
			if j&s == 0 {
				states[j] = e.states[j|s]
			} else {
				states[j] = e.states[j|s|t]
			}
		}
		e.states = states
		if i != m {
			e.qubits[i] = e.qubits[m]
			e.qubits[i].i = i
		}
		e.qubits = e.qubits[:m]
		return 1
	}
}
