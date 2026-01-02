package blocker

import "adblocker/internal/blocklist"

// StandardBlocker implementa a interface Blocker usando uma lista de domínios em memória.
type StandardBlocker struct {
	matcher *blocklist.Matcher
}

// NewStandardBlocker cria uma nova instância do StandardBlocker com um Matcher vazio.
func NewStandardBlocker() *StandardBlocker {
	return &StandardBlocker{
		matcher: blocklist.NewMatcher(),
	}
}

// LoadFromFile lê um arquivo de lista de bloqueio e adiciona os domínios.
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

// LoadFromURL lê uma lista de uma URL e adiciona os domínios.
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

// AddRegex adiciona uma regra de bloqueio baseada em expressão regular.
func (b *StandardBlocker) AddRegex(pattern string) error {
	return b.matcher.AddRegex(pattern)
}

// Count retorna o total de domínios carregados.
func (b *StandardBlocker) Count() int {
	return b.matcher.Count()
}

// ShouldBlock retorna true se o domínio (ou um pai dele) estiver na lista.
func (b *StandardBlocker) ShouldBlock(domain string) bool {
	return b.matcher.Match(domain)
}