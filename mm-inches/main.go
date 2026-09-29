package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed web/templates web/static
var webFiles embed.FS

type server struct {
	index *template.Template
}

type conversionRequest struct {
	Unit                string `json:"unit"`
	Value               string `json:"value"`
	DecimalPlaces       int    `json:"decimalPlaces"`
	FractionDenominator int    `json:"fractionDenominator"`
}

type conversionResponse struct {
	Values map[string]string `json:"values"`
}

func main() {
	tmpl := template.Must(template.ParseFS(webFiles, "web/templates/index.html"))
	staticFiles, err := fs.Sub(webFiles, "web/static")
	if err != nil {
		log.Fatal(err)
	}

	s := &server{index: tmpl}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /api/convert", s.handleConvert)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("unit converter listening on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.index.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "unable to render page", http.StatusInternalServerError)
	}
}

func (s *server) handleConvert(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	defer r.Body.Close()

	var req conversionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, "The conversion request was not valid.", http.StatusBadRequest)
		return
	}

	if req.DecimalPlaces < 0 || req.DecimalPlaces > 6 {
		writeJSONError(w, "Decimal accuracy must be between 0 and 6 places.", http.StatusBadRequest)
		return
	}
	if !validDenominator(req.FractionDenominator) {
		writeJSONError(w, "Fraction accuracy must be 1/2, 1/4, 1/8, 1/16, 1/32, or 1/64.", http.StatusBadRequest)
		return
	}

	value, err := parseValue(req.Unit, req.Value)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	values, err := convertFrom(req.Unit, value, req.DecimalPlaces, req.FractionDenominator)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conversionResponse{Values: values})
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func parseValue(unit, raw string) (float64, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, fmt.Errorf("Enter a value to convert")
	}
	if unit == "fractional-inches" {
		value, err := parseImperialFraction(raw)
		if err != nil {
			return 0, fmt.Errorf("Use a fraction such as 16 3/4")
		}
		return value, nil
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(raw, ",", "")), 64)
	if err != nil {
		return 0, fmt.Errorf("Enter a valid number")
	}
	return value, nil
}
