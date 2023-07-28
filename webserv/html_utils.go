package webserv

// Escaper interface performs html escaping on its argument
type Escaper interface {
	Escaped() string
}

