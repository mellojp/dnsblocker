package blocker

type Blocker interface {
	ShouldBlock(domain string) bool
}
