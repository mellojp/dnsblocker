package blocker

import "dnsblocker/internal/blocklist"

type StandardBlocker struct {
	matcher *blocklist.Matcher
}

func NewStandardBlocker() *StandardBlocker {
	return &StandardBlocker{
		matcher: blocklist.NewMatcher(),
	}
}

func (b *StandardBlocker) LoadFromFile(path string) error {
	domains, err := blocklist.LoadFromFile(path)
	if err != nil {
		return err
	}
	for _, d := range domains {
		b.matcher.Add(d)
	}
	return nil
}

func (b *StandardBlocker) LoadFromURL(url string) error {
	domains, err := blocklist.LoadFromURL(url)
	if err != nil {
		return err
	}
	for _, d := range domains {
		b.matcher.Add(d)
	}
	return nil
}

func (b *StandardBlocker) AddRegex(pattern string) error {
	return b.matcher.AddRegex(pattern)
}

func (b *StandardBlocker) Count() int {
	return b.matcher.Count()
}

func (b *StandardBlocker) ShouldBlock(domain string) bool {
	return b.matcher.Match(domain)
}
