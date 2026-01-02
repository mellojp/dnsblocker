package dns

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

// itemCache representa uma resposta DNS armazenada
type itemCache struct {
	msg        *dns.Msg
	expiration time.Time
}

// MemoryCache gerencia o armazenamento de respostas DNS em memória RAM
type MemoryCache struct {
	items map[string]itemCache
	mu    sync.RWMutex
}

// NewMemoryCache cria um novo cache vazio
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]itemCache),
	}
}

// Get busca uma resposta no cache.
// Retorna a mensagem e 'true' se encontrar e não estiver expirada.
func (c *MemoryCache) Get(key string) (*dns.Msg, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Verifica se já expirou (TTL acabou)
	if time.Now().After(item.expiration) {
		return nil, false
	}

	// Retorna uma cópia da mensagem para evitar concorrência na modificação do ID depois
	return item.msg.Copy(), true
}

// Set salva uma resposta no cache
func (c *MemoryCache) Set(key string, msg *dns.Msg) {
	if len(msg.Answer) == 0 {
		return // Não cacheia respostas vazias
	}

	// Calcula o TTL (Time To Live) baseado na primeira resposta
	// O padrão é usar o menor TTL encontrado nos registros para ser seguro
	ttl := time.Duration(msg.Answer[0].Header().Ttl) * time.Second

	// Cria uma cópia para salvar
	toSave := msg.Copy()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = itemCache{
		msg:        toSave,
		expiration: time.Now().Add(ttl),
	}
}

// MakeKey gera uma chave única para a pergunta (Nome + Tipo)
// Ex: "google.com.|1" (1 = Tipo A)
func MakeKey(q dns.Question) string {
	return q.Name + "|" + string(rune(q.Qtype))
}
