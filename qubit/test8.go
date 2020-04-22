package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var Clock int

type Gate []complex64

var sqrt5 float32 = 1 / math.Sqrt2
var (
	H    = Gate{complex(sqrt5, 0), complex(sqrt5, 0), complex(sqrt5, 0), complex(-sqrt5, 0)}
	X    = Gate{0, 1, 1, 0}
	Not  = X
	Y    = Gate{0, -1i, 1i, 0}
	Z    = Gate{1, 0, 0, -1}
	S    = Gate{1, 0, 0, 1i}
	InvS = Gate{1, 0, 0, -1i}
	T    = Gate{1, 0, 0, complex(sqrt5, sqrt5)}
	InvT = Gate{1, 0, 0, complex(sqrt5, -sqrt5)}
)

/*func (q *Qubit) String() string {
	q.init()
	if len(q.qs) > 2 {
		return "???? (entangled)"
	} else {
		a := complex128(q.qs[0])
		b := complex128(q.qs[1])
		s := cmplx.Sqrt(a*-a + b*-b)
		a /= s
		b /= s
		return fmt.Sprintf("%.3f|0> + %.3f|1> (unentangled)", a, b)
	}
}*/

func (q *qubit) String() string {
	if len(q.e.qubits) == 1 {
		if q.e.states[1] == 0 {
			return "|0>"
		}
		if q.e.states[0] == 0 {
			return "|1>"
		}
		return "|?> (unentangled)"
	} else {
		return fmt.Sprintf("|?> (entangled %d)", len(q.e.qubits))
	}
}

type qubit struct {
	e *entanglement
	i int
}

func NewQubit() *qubit {
	q := &qubit{}
	q.e = &entanglement{
		states: []complex64{1, 0},
		qubits: []*qubit{q},
	}
	return q
}

func (q *qubit) Swap(other *qubit) {
	*q, *other = *other, *q
}

func (q *qubit) apply(gate Gate) {
	q.e.apply(q.i, gate)
}

func (q *qubit) control(o *qubit, gate Gate) {
	if len(q.e.qubits) == 1 {
		if q.e.states[1] == 0 {
			Clock++
			return
		}
		if q.e.states[0] == 0 {
			o.apply(gate)
			return
		}
	}
	q.e.join(o.e)
	q.e.control(q.i, o.i, gate)
}

func (q *qubit) measure() int {
	return q.e.measure(q.i)
}

func toffoli(a, b, c *qubit) {
	c.apply(H)
	b.control(c, S)
	a.control(b, X)
	b.control(c, InvS)
	a.control(b, X)
	a.control(c, S)
	c.apply(H)
}

func xor(a, b, c *qubit) {
	a.control(c, X)
	b.control(c, X)
}

func or(a, b, c *qubit) {
	a.apply(X)
	b.apply(X)
	toffoli(a, b, c)
	a.apply(X)
	b.apply(X)
	c.apply(X)
}

func output(bs []*qubit) {
	for i := len(bs) - 1; i >= 0; i-- {
		fmt.Printf("%d ", bs[i].measure())
	}
	/*for i := range bs {
		fmt.Printf("%d ", bs[i].measure())
	}*/
	fmt.Printf("\n")
}
func and(a, b, c *qubit) {
	toffoli(a, b, c)
}

func zero(a *qubit) {
	i := a.measure()
	if i == 1 {
		a.apply(X)
	}
}

type Qureg []*qubit

func NewQureg(size int) Qureg {
	r := make(Qureg, size)
	for i := range r {
		r[i] = NewQubit()
	}
	return r
}

func output2(a Qureg) {
	ent := a[0].e
	var ff complex64
	var zeroes, nonzeroes int
	for _, state := range ent.states {
		if state == 0 {
			zeroes++
		} else {
			if state != ff {
				fmt.Print(state, " ")
				ff = state
			}
			nonzeroes++
		}
	}
	fmt.Println()
	fmt.Printf("ZEROES=%d NONZEROES=%d\n", zeroes, nonzeroes)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	LEN := 8 //8 to cirka 30 sek - nu cirka 6 sek når man measurer inden man er færdig
	a := NewQureg(LEN)
	b := NewQureg(LEN)
	c := NewQureg(LEN)

	for i := range a {
		a[i].apply(H)
		b[i].apply(H)
	}

	d := NewQubit()
	e := NewQubit()
	f := NewQubit()
	for i := range a {
		if i+1 != len(a) {
			and(a[i], b[i], d)
			and(a[i], c[i], e)
			or(d, e, f)
			and(a[i], b[i], d)
			and(a[i], c[i], e)

			and(b[i], c[i], e)
			or(e, f, c[i+1])
			and(b[i], c[i], e)

			and(a[i], b[i], d)
			and(a[i], c[i], e)
			or(d, e, f)
			and(a[i], b[i], d)
			and(a[i], c[i], e)
			f.measure()
		}
		xor(a[i], b[i], c[i])
		fmt.Println(Clock)
	}
	/*for i := range a {
		if i+1 != len(a) {
			d := NewQubit()
			e := NewQubit()
			and(a[i], b[i], d)
			and(a[i], c[i], e)
			f := NewQubit()
			or(d, e, f)
			d.measure()
			e.measure()
			and(b[i], c[i], e)
			or(e, f, c[i+1])
			e.measure()
			f.measure()
		}
		xor(a[i], b[i], c[i])
		fmt.Println(Clock)
	}*/
	output2(a)

	output(a)
	output(b)
	output(c)

	/*bs := make([]*qubit, 25)
	for i := range bs {
		bs[i] = new(qubit)
		fmt.Printf("%d\n", i)
		if i == 0 {
			bs[0].apply(Hadamard)
		} else {
			bs[0].control(bs[i], PauliX)
		}
	}
	output(bs)
	*/
}
