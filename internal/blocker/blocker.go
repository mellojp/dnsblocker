package blocker

// Blocker define o comportamento de qualquer estratégia de bloqueio.
type Blocker interface {
	ShouldBlock(domain string) bool
}
