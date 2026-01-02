package blocklist

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const cacheTTL = 24 * time.Hour

// ParseReader é uma função auxiliar genérica que lê de qualquer fonte de dados (arquivo ou rede)
// e extrai os domínios, limpando IPs (formato HOSTS) e comentários.
func ParseReader(r io.Reader) ([]string, error) {
	var domains []string
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		// Remove comentários inline (ex: "google.com # comentario")
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}

		// Remove espaços nas pontas
		line = strings.TrimSpace(line)

		// Ignora linhas vazias
		if line == "" {
			continue
		}

		// Processa formato HOSTS (ex: "0.0.0.0 domain.com")
		// Divide a linha por espaços
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		var domain string

		// Se tiver mais de 1 campo, assumimos que é formato HOSTS (IP DOMINIO)
		// Pegamos o segundo campo.
		if len(fields) >= 2 {
			domain = fields[1]
		} else {
			// Se tiver só 1 campo, assumimos que é uma lista simples de domínios
			domain = fields[0]
		}

		// Normalizações finais
		domain = strings.ToLower(domain)
		domain = strings.TrimSuffix(domain, ".")

		// Filtro básico: se o domínio for "localhost" ou IPs locais, ignoramos
		if domain == "localhost" || domain == "255.255.255.255" || domain == "broadcasthost" {
			continue
		}

		domains = append(domains, domain)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

// LoadFromFile usa a função genérica para carregar de um arquivo local.
func LoadFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseReader(file)
}

// LoadFromURL baixa uma lista da internet e extrai os domínios.
func LoadFromURL(url string) ([]string, error) {
	// 1. Gera um nome de arquivo baseado no hash da URL
	hash := md5.Sum([]byte(url))
	fileName := hex.EncodeToString(hash[:]) + ".txt"
	cachePath := filepath.Join("cache", fileName)

	// 2. Verifica se o arquivo existe e se é recente
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			// O arquivo está no cache e é recente. Vamos usar ele!
			return LoadFromFile(cachePath)
		}
	}

	// 3. Se não existe ou expirou, fazemos o download
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		// Se o download falhar mas tivermos um cache antigo, usamos o antigo como fallback
		if _, errFile := os.Stat(cachePath); errFile == nil {
			return LoadFromFile(cachePath)
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao baixar lista: status %d", resp.StatusCode)
	}

	// 4. Salva no disco ANTES de retornar para persistência
	out, err := os.Create(cachePath)
	if err == nil {
		// TEE Reader: Lê da rede e escreve no arquivo simultaneamente
		_, _ = io.Copy(out, resp.Body)
		out.Close()
	}

	// 5. Agora lê o arquivo salvo para extrair os domínios
	return LoadFromFile(cachePath)
}
