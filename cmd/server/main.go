package main

import (
	"dnsblocker/internal/blocker"
	"dnsblocker/internal/config"
	"dnsblocker/internal/dashboard"
	"dnsblocker/internal/dns"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetOutput(os.Stdout)

	if err := os.MkdirAll("cache", 0755); err != nil {
		log.Printf("Aviso: Não foi possível criar pasta de cache: %v", err)
	}

	iface, err := config.GetPrimaryInterface()
	if err == nil {
		log.Printf("Alterando DNS da interface '%s' para 127.0.0.1...", iface)
		if err := config.SetSystemDNS(iface); err != nil {
			log.Panicf("ERRO ao alterar DNS: %v", err)
		}

		defer func() {
			log.Printf("Restaurando DNS da interface '%s'...", iface)
			config.RestoreDNS(iface)
		}()
	}

	myBlocker := blocker.NewStandardBlocker()
	logger := dashboard.NewLogBuffer()

	for _, pattern := range config.RegexRules {
		if err := myBlocker.AddRegex(pattern); err != nil {
			log.Printf("ERRO ao adicionar regex '%s': %v", pattern, err)
		}
	}

	log.Println("Carregando listas de bloqueio (usando cache local se disponível)...")
	for _, url := range config.BlocklistURLs {
		if err := myBlocker.LoadFromURL(url); err != nil {
			log.Printf("ERRO: Falha ao carregar %s: %v", url, err)
		}
	}
	log.Printf("Processamento de listas finalizado. Total de domínios na lista de bloqueio: %d", myBlocker.Count())

	server := dns.NewDNSServer(config.ListenAddr, config.UpstreamDNS, myBlocker, logger)

	go func() {
		app := dashboard.NewDashApp(logger, func() (uint64, uint64, uint64) {
			return server.Stats.Total.Load(),
				server.Stats.Blocked.Load(),
				server.Stats.Cached.Load()
		})
		if err := app.Start(":8080"); err != nil {
			log.Printf("Erro no dashboard: %v", err)
		}
	}()

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("ERRO CRÍTICO: Falha ao iniciar servidor DNS: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\nRecebido sinal de encerramento. Parando servidor...")
}
