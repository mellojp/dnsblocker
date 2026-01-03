package dns

import (
	"dnsblocker/internal/blocker"
	"dnsblocker/internal/dashboard"
	"log"
	"sync/atomic"

	"github.com/miekg/dns"
)

type Server struct {
	addr     string
	resolver *Resolver
	blocker  blocker.Blocker
	cache    *MemoryCache
	logger   *dashboard.LogBuffer
	Stats    struct {
		Total   atomic.Uint64
		Blocked atomic.Uint64
		Cached  atomic.Uint64
	}
}

func NewDNSServer(addr string, upstream string, blocker blocker.Blocker, logBuf *dashboard.LogBuffer) *Server {
	return &Server{
		addr:     addr,
		resolver: NewResolver(upstream),
		blocker:  blocker,
		cache:    NewMemoryCache(),
		logger:   logBuf,
	}
}

func (s *Server) Start() error {
	dns.HandleFunc(".", s.handleDNSRequest)

	errChan := make(chan error)

	go func() {
		serverUDP := &dns.Server{
			Addr:    s.addr,
			Net:     "udp",
			UDPSize: 65535,
		}
		log.Printf("DNS server ouvindo em %s/udp", s.addr)
		if err := serverUDP.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		serverTCP := &dns.Server{Addr: s.addr, Net: "tcp"}
		log.Printf("DNS server ouvindo em %s/tcp", s.addr)
		if err := serverTCP.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	return <-errChan
}

func (s *Server) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		return
	}

	s.Stats.Total.Add(1)
	q := r.Question[0]
	domain := q.Name

	checkDomain := domain
	if len(checkDomain) > 0 && checkDomain[len(checkDomain)-1] == '.' {
		checkDomain = checkDomain[:len(checkDomain)-1]
	}

	if s.blocker.ShouldBlock(checkDomain) {
		s.Stats.Blocked.Add(1)
		str := `<span class="tag-blocked">[BLOCKED]</span> ` + checkDomain
		s.logger.AddLog(str)
		log.Printf("[\033[1;33mBLOCKED\033[0m] %s", checkDomain)
		msg := new(dns.Msg)
		msg.SetReply(r)

		rr, _ := dns.NewRR(q.Name + " 3600 IN A 0.0.0.0")
		msg.Answer = append(msg.Answer, rr)

		w.WriteMsg(msg)
		return
	}

	cacheKey := MakeKey(q)
	if cachedMsg, found := s.cache.Get(cacheKey); found {
		s.Stats.Cached.Add(1)
		str := `<span class="tag-cached">[CACHED]</span> ` + checkDomain
		s.logger.AddLog(str)
		log.Printf("[\033[1;36mCACHED\033[0m] %s", checkDomain)
		cachedMsg.Id = r.Id
		cachedMsg.Compress = true
		w.WriteMsg(cachedMsg)
		return
	}

	resp, err := s.resolver.Forward(r)
	if err != nil {
		str := `<span class="tag-error">[ERROR]</span> ` + checkDomain
		s.logger.AddLog(str)
		log.Printf("\033[1;31m[ERROR]\033[0m %s: %v", checkDomain, err)
		fail := new(dns.Msg)
		fail.SetReply(r)
		fail.Rcode = dns.RcodeServerFailure
		w.WriteMsg(fail)
		return
	}

	str := `<span class="tag-allowed">[ALLOW]</span> ` + checkDomain
	s.logger.AddLog(str)
	log.Printf("[\033[1;32mALLOW\033[0m] %s", checkDomain)

	s.cache.Set(cacheKey, resp)

	resp.Id = r.Id
	resp.Compress = true

	w.WriteMsg(resp)
}
