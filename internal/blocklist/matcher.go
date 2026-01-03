package blocklist

import (
	"regexp"
	"strings"
)

type Matcher struct {
	blocked map[string]struct{}
	regexes []*regexp.Regexp
}

func NewMatcher() *Matcher {
	return &Matcher{
		blocked: make(map[string]struct{}),
		regexes: make([]*regexp.Regexp, 0),
	}
}

func (m *Matcher) Add(domain string) {
	m.blocked[strings.ToLower(domain)] = struct{}{}
}

func (m *Matcher) AddRegex(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	m.regexes = append(m.regexes, re)
	return nil
}

func (m *Matcher) Count() int {
	return len(m.blocked)
}

func (m *Matcher) Match(domain string) bool {
	if domain == "" {
		return false
	}

	current := strings.ToLower(domain)

	temp := current
	for {
		if _, exists := m.blocked[temp]; exists {
			return true
		}

		firstDot := strings.IndexByte(temp, '.')
		if firstDot == -1 {
			break
		}

		temp = temp[firstDot+1:]

		if temp == "" {
			break
		}
	}

	for _, re := range m.regexes {
		if re.MatchString(current) {
			return true
		}
	}

	return false
}
