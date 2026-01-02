package config

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	ListenAddr  string = ":53"
	UpstreamDNS string = "8.8.8.8:53"
)

var BlocklistURLs = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://big.oisd.nl",
	"https://raw.githubusercontent.com/badmojr/1Hosts/master/Lite/hosts.txt",
	"https://raw.githubusercontent.com/anudeepND/blacklist/master/adservers.txt",
}

var RegexRules = []string{
	`^rr[\w-]+\.googlevideo\.com$`, // Ex: rr1---sn-b8u-5w0e.googlevideo.com
	`^ad[s]?[\w-]*\.`,              // Ex: ads-api.twitter.com, adserver.com
}

// GetPrimaryInterface tenta encontrar o nome da interface de rede ativa (ex: "Wi-Fi" ou "Ethernet")
func GetPrimaryInterface() (string, error) {
	// Comando para listar interfaces conectadas
	out, err := exec.Command("netsh", "interface", "show", "interface").Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Connected") || strings.Contains(line, "Conectado") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[len(fields)-1], nil
			}
		}
	}
	return "", fmt.Errorf("nenhuma interface ativa encontrada")
}

// SetSystemDNS altera o DNS da interface para local
func SetSystemDNS(iface string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns", iface, "static", "127.0.0.1")
	return cmd.Run()
}

// RestoreDNS volta o DNS para automático (DHCP)
func RestoreDNS(iface string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns", iface, "dhcp")
	return cmd.Run()
}
