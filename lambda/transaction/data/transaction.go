package data

type Transaction struct {
	Name     string
	Base     int
	Fraction int
	Note     string
}

type Amount struct {
	Base     int // e.g. eur, usd
	Fraction int // e.g. cents
}
