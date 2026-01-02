package blocklist

import (
	"regexp"
	"strings"
)

// Matcher é a estrutura responsável por armazenar a lista de domínios bloqueados
// e verificar eficientemente se um dado domínio (ou seus subdomínios) está na lista.
type Matcher struct {
	blocked map[string]struct{}
	regexes []*regexp.Regexp
}

// NewMatcher cria uma nova instância de Matcher vazia.
func NewMatcher() *Matcher {
	return &Matcher{
		blocked: make(map[string]struct{}),
		regexes: make([]*regexp.Regexp, 0),
	}
}

// Add adiciona um domínio à lista de bloqueio.
func (m *Matcher) Add(domain string) {
	// Armazenamos sempre em minúsculas para garantir consistência
	m.blocked[strings.ToLower(domain)] = struct{}{}
}

// AddRegex adiciona uma expressão regular à lista de bloqueio.
func (m *Matcher) AddRegex(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	m.regexes = append(m.regexes, re)
	return nil
}

// Count retorna o número total de domínios na lista (apenas exatos).
func (m *Matcher) Count() int {
	return len(m.blocked)
}

// Match verifica se o domínio ou qualquer um de seus domínios pai está na lista.
// Exemplo: Se "example.com" estiver bloqueado, "ads.example.com" retornará true.
// Também verifica se o domínio corresponde a alguma das expressões regulares cadastradas.
func (m *Matcher) Match(domain string) bool {
	if domain == "" {
		return false
	}

	// Normaliza para minúsculas antes de verificar
	current := strings.ToLower(domain)

	// 1. Verificação Hierárquica (Exact Match & Parents)
	// Loop para verificar a hierarquia do domínio
	// Ex: "ads.google.com" -> verifica "ads.google.com" -> verifica "google.com" -> verifica "com"
	temp := current
	for {
		if _, exists := m.blocked[temp]; exists {
			return true
		}

		// Encontra o próximo ponto para remover o subdomínio mais à esquerda
		firstDot := strings.IndexByte(temp, '.')
		if firstDot == -1 {
			// Não há mais pontos, chegamos ao final do domínio
			break
		}

		// Avança para a próxima parte do domínio (remove tudo antes do primeiro ponto)
		temp = temp[firstDot+1:]

		// Se a string ficou vazia, paramos
		if temp == "" {
			break
		}
	}

	// 2. Verificação por Regex (Heurísticas)
	for _, re := range m.regexes {
		if re.MatchString(current) {
			return true
		}
	}

	return false
}
