package main

type Retrier struct {
	Interval time.Interval
	Multiplier float64
	Randomness float64
	Context context.Context
}

func (r* Retrier) Next() error {
	interval := r.Interval
	r.Interval *= r.Multiplier
	interval *= 1.0 - r.Randomness + 2 * r.Randomness * rand.Float64()
	t := time.NewTimer(interval)
	select {
	case <-t.C:
	case <-r.Context.Done():
		t.Stop()
	}
	return r.Context.Err()
}
