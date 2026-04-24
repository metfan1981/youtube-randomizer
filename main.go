package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	// Don't require a readable .env for -h / -help (education + broken local files).
	if !argsContainHelpFlag(os.Args[1:]) {
		if err := loadDotEnv(".env"); err != nil {
			fmt.Fprintf(os.Stderr, "load .env: %v\n", err)
			os.Exit(1)
		}
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Serves random YouTube uploads for /@handle paths.\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  YT_API_KEY    required — YouTube Data API key\n")
		fmt.Fprintf(os.Stderr, "  HOST          optional — listen IP if -host is not set (default 127.0.0.1)\n")
		fmt.Fprintf(os.Stderr, "  PORT          optional — listen port if -port is not set (default 8080)\n")
		fmt.Fprintf(os.Stderr, "  CACHE_TTL     optional — upload list cache TTL (default 1h)\n")
		fmt.Fprintf(os.Stderr, "  FETCH_TIMEOUT optional — max time for one channel list fetch (default 3m)\n")
		fmt.Fprintf(os.Stderr, "\nIf .env exists in the working directory, it sets missing variables only\n")
		fmt.Fprintf(os.Stderr, "(systemd: use EnvironmentFile= — WorkingDirectory is usually not the repo).\n")
	}

	// Explicit -h/-help exits 0 (stdlib default is exit 2); keep both for clarity.
	exitUsage := func(_ string) error {
		flag.Usage()
		os.Exit(0)
		return nil
	}
	flag.BoolFunc("help", "print help and exit", exitUsage)
	flag.BoolFunc("h", "print help and exit (shorthand)", exitUsage)

	host := flag.String("host", "", "listen `IP` or hostname (default $HOST or 127.0.0.1; use 0.0.0.0 for all interfaces)")
	portFlag := flag.String("port", "", "listen `port` (overrides $PORT; default 8080)")
	flag.Parse()

	hostStr := *host
	if hostStr == "" {
		hostStr = os.Getenv("HOST")
	}
	if hostStr == "" {
		hostStr = "127.0.0.1"
	}

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	addr := net.JoinHostPort(hostStr, port)

	apiKey := os.Getenv("YT_API_KEY")
	if apiKey == "" {
		log.Error("YT_API_KEY is required")
		os.Exit(1)
	}

	ttlStr := os.Getenv("CACHE_TTL")
	if ttlStr == "" {
		ttlStr = "1h"
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		log.Error("CACHE_TTL", "err", err)
		os.Exit(1)
	}

	fetchStr := os.Getenv("FETCH_TIMEOUT")
	if fetchStr == "" {
		fetchStr = "3m"
	}
	fetchTimeout, err := time.ParseDuration(fetchStr)
	if err != nil {
		log.Error("FETCH_TIMEOUT", "err", err)
		os.Exit(1)
	}

	yt, err := newClient(apiKey, ttl)
	if err != nil {
		log.Error("youtube client", "err", err)
		os.Exit(1)
	}

	h, err := newHandler(yt, log, fetchTimeout)
	if err != nil {
		log.Error("handler", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	// Exact "/" only; plain "GET /" is a prefix match and would serve /foo as home.
	mux.HandleFunc("GET /{$}", h.home)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	// One path segment, e.g. /@mkbhd — "@{handle}" is invalid in Go 1.22+ patterns.
	mux.HandleFunc("GET /{handle}", h.channel)

	// WriteTimeout must cover handler work (YouTube pagination on cache miss) plus response write.
	writeTimeout := fetchTimeout + 60*time.Second
	if writeTimeout < 90*time.Second {
		writeTimeout = 90 * time.Second
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       120 * time.Second,
	}
	log.Info("listening", "host", hostStr, "port", port, "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}
