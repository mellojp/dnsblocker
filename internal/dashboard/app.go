package dashboard

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

//go:embed pages/*
var webFiles embed.FS

type DashApp struct {
	buffer   *LogBuffer
	getStats func() (uint64, uint64, uint64)
}

func NewDashApp(b *LogBuffer, statsFunc func() (uint64, uint64, uint64)) *DashApp {
	return &DashApp{
		buffer:   b,
		getStats: statsFunc,
	}
}

func (app *DashApp) Start(port string) error {
	// Apenas duas rotas principais e uma para atualização de stats
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/api/stats", app.handleStatsAPI)
	http.HandleFunc("/events", app.handleEvents)

	fmt.Printf("Dashboard consolidado rodando em http://localhost%s\n", port)
	return http.ListenAndServe(port, nil)
}

func (app *DashApp) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(webFiles, "pages/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	total, blocked, cached := app.getStats()
	data := struct {
		Logs    []string
		Total   uint64
		Blocked uint64
		Cached  uint64
		Percent float64
	}{
		Logs:    app.buffer.GetLogs(),
		Total:   total,
		Blocked: blocked,
		Cached:  cached,
	}
	if total > 0 {
		data.Percent = (float64(blocked) / float64(total)) * 100
	}

	tmpl.Execute(w, data)
}

// Endpoint para o HTMX atualizar apenas os cards
func (app *DashApp) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	total, blocked, cached := app.getStats() // Agora pegamos os 3 valores
	percent := 0.0
	if total > 0 {
		percent = (float64(blocked) / float64(total)) * 100
	}

	// Retorna o HTML dos cards que será injetado no #stats-grid
	fmt.Fprintf(w, `
		<div class="stat-card"><span>%d</span>Total</div>
		<div class="stat-card"><span style="color: #f44747;">%d</span>Bloqueios</div>
		<div class="stat-card"><span style="color: #4ec9b0;">%.1f%%</span>Porcentagem bloqueada</div>
		<div class="stat-card"><span style="color: #4fc1ff;">%d</span>Cache</div>
		
	`, total, blocked, percent, cached)
}

func (app *DashApp) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	logChan := app.buffer.AddListener()

	for {
		select {
		case msg := <-logChan:
			tipo := "unknown"
			if strings.Contains(msg, "tag-blocked") {
				tipo = "blocked"
			}
			if strings.Contains(msg, "tag-allowed") {
				tipo = "allowed"
			}
			if strings.Contains(msg, "tag-cached") {
				tipo = "cached"
			}

			// Capturamos o timestamp atual em segundos
			ts := time.Now().Unix()

			// Adicionamos o atributo data-ts para o JavaScript usar
			fmt.Fprintf(w, "data: <div class='log-line' data-type='%s' data-ts='%d'>%s</div>\n\n",
				tipo, ts, msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}
