package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed all:dist
var dist embed.FS

type headline struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Source      string `json:"source"`
	Summary     string `json:"summary"`
	PublishedAt string `json:"publishedAt"`
}

type headlinesResponse struct {
	Items []headline `json:"items"`
}

type metarResponse struct {
	ICAO       string `json:"icao"`
	Raw        string `json:"raw"`
	Wind       string `json:"wind"`
	Visibility string `json:"visibility"`
	Sky        string `json:"sky"`
}

func headlinesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	now := time.Now().UTC()
	items := []headline{
		{
			ID:          "1",
			Title:       "EASA publishes winter ops reminder for contaminated runways",
			Source:      "Ops desk",
			Summary:     "Operators reminded to validate braking action reports and crosswind limits when slush or compacted snow is present.",
			PublishedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:          "2",
			Title:       "North Atlantic tracks shift south for jet stream",
			Source:      "Planning",
			Summary:     "Track structure adjusted through the weekend; check master weather package before filing.",
			PublishedAt: now.Add(-5 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:          "3",
			Title:       "Airport curfew window shortened at select EU hubs",
			Source:      "Notices",
			Summary:     "Temporary noise abatement changes affect night slots; confirm with your handler.",
			PublishedAt: now.Add(-26 * time.Hour).Format(time.RFC3339),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(headlinesResponse{Items: items})
}

func metarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	icao := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("icao")))
	if icao == "" {
		icao = "KSFO"
	}
	// Demo payload — not live aviation weather.
	body := metarResponse{
		ICAO:       icao,
		Raw:        icao + " 091200Z 28014G22KT 10SM FEW025 SCT250 12/M04 A2992 RMK AO2",
		Wind:       "280° at 14 kt, gusts 22 kt",
		Visibility: "10 SM",
		Sky:        "FEW025 SCT250",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"ready": "true"})
}

func spaHandler(static fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(static))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		upath := r.URL.Path
		if upath == "/" {
			upath = "/index.html"
		}
		name := strings.TrimPrefix(upath, "/")
		if strings.Contains(name, "..") {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		if _, err := fs.Stat(static, name); err != nil {
			upath = "/index.html"
		}
		r2 := *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = upath
		fileServer.ServeHTTP(w, &r2)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/headlines", headlinesHandler)
	mux.HandleFunc("/api/metar", metarHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.Handle("/", spaHandler(sub))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
