package blocker

type AllowAllBlocker struct{}

func NewAllowAllBlocker() *AllowAllBlocker {
	return &AllowAllBlocker{}
}

func (b *AllowAllBlocker) ShouldBlock(domain string) bool {
	return false
}
