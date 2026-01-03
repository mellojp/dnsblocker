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

func ParseReader(r io.Reader) ([]string, error) {
	var domains []string
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		var domain string

		if len(fields) >= 2 {
			domain = fields[1]
		} else {
			domain = fields[0]
		}

		domain = strings.ToLower(domain)
		domain = strings.TrimSuffix(domain, ".")

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

func LoadFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseReader(file)
}

func LoadFromURL(url string) ([]string, error) {
	hash := md5.Sum([]byte(url))
	fileName := hex.EncodeToString(hash[:]) + ".txt"
	cachePath := filepath.Join("cache", fileName)

	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			return LoadFromFile(cachePath)
		}
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		if _, errFile := os.Stat(cachePath); errFile == nil {
			return LoadFromFile(cachePath)
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao baixar lista: status %d", resp.StatusCode)
	}

	out, err := os.Create(cachePath)
	if err == nil {
		_, _ = io.Copy(out, resp.Body)
		out.Close()
	}

	return LoadFromFile(cachePath)
}
