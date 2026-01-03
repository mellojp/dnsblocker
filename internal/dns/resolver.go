package dns

import (
	"log"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
	upstream  string
	udpClient *dns.Client
	tcpClient *dns.Client
}

func NewResolver(upstream string) *Resolver {
	return &Resolver{
		upstream: upstream,
		udpClient: &dns.Client{
			Net:     "udp",
			Timeout: 500 * time.Millisecond,
		},
		tcpClient: &dns.Client{
			Net:     "tcp",
			Timeout: 5 * time.Second,
		},
	}
}

func (r *Resolver) Forward(msg *dns.Msg) (*dns.Msg, error) {
	start := time.Now()

	resp, _, err := r.udpClient.Exchange(msg, r.upstream)

	if err != nil || (resp != nil && resp.Truncated) {
		log.Printf("[QUERY] %s", msg.Question[0].Name)
		log.Printf("[Resolver] UDP falhou (err=%v, trunc=%v) após %v. Tentando TCP...", err, (resp != nil && resp.Truncated), time.Since(start))

		respTCP, _, errTCP := r.tcpClient.Exchange(msg, r.upstream)
		log.Printf("[Resolver] TCP finalizado em %v (Total: %v)", time.Since(start), time.Since(start))
		return respTCP, errTCP
	}

	return resp, nil
}
