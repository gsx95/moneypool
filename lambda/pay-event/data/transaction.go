package data

import "fmt"

type Transaction struct {
	Name     string
	Base     int
	Fraction int
	Note     string
}

func (t Transaction) AmountString() string {
	return fmt.Sprintf("%d.%d", t.Base, t.Fraction)
}

type Amount struct {
	Base     int // e.g. eur, usd
	Fraction int // e.g. cents
}
