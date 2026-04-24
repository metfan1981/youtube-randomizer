package main

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//go:embed page.html
var pageFS embed.FS

// YouTube handles are 3–30 chars (alphanumeric, _, ., -).
var handleRe = regexp.MustCompile(`^[a-z0-9._-]{3,30}$`)

// Bare handle names that look like HTTP junk; avoids a YouTube round-trip for /@favicon.ico.
var blockedBareHandles = map[string]struct{}{
	"favicon.ico": {},
	"robots.txt":  {},
}

type Handler struct {
	yt           *Client
	log          *slog.Logger
	tmpl         *template.Template
	fetchTimeout time.Duration
}

func newHandler(yt *Client, log *slog.Logger, fetchTimeout time.Duration) (*Handler, error) {
	tmpl, err := template.ParseFS(pageFS, "page.html")
	if err != nil {
		return nil, err
	}
	return &Handler{yt: yt, log: log, tmpl: tmpl, fetchTimeout: fetchTimeout}, nil
}

type pageData struct {
	Handle string
	Videos []string
	Count  int
}

func (h *Handler) channel(w http.ResponseWriter, r *http.Request) {
	seg := strings.ToLower(r.PathValue("handle"))
	if !strings.HasPrefix(seg, "@") {
		http.Error(w, "invalid handle", http.StatusBadRequest)
		return
	}
	bare := strings.TrimPrefix(seg, "@")
	if _, junk := blockedBareHandles[bare]; junk {
		http.Error(w, "invalid handle", http.StatusBadRequest)
		return
	}
	if !handleRe.MatchString(bare) {
		http.Error(w, "invalid handle", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.fetchTimeout)
	defer cancel()

	ids, err := h.yt.Uploads(ctx, seg)
	if err != nil {
		h.log.Error("uploads", "err", err)
		http.Error(w, "could not load channel", http.StatusBadGateway)
		return
	}
	if len(ids) == 0 {
		http.Error(w, "no uploads found", http.StatusNotFound)
		return
	}

	n := min(3, len(ids))
	perm := rand.Perm(len(ids))
	chosen := make([]string, n)
	for i := range n {
		chosen[i] = ids[perm[i]]
	}

	var buf bytes.Buffer
	data := pageData{Handle: bare, Videos: chosen, Count: n}
	if err := h.tmpl.Execute(&buf, data); err != nil {
		h.log.Error("template", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>youtube-randomizer</title></head>
<body style="font-family:system-ui;max-width:40rem;margin:2rem auto;line-height:1.5">
<p>Open a channel by handle in the path, for example:</p>
<pre style="background:#f4f4f4;padding:1rem">/<wbr>@mkbhd</pre>
<p>Refresh the page to pick new random uploads (up to three).</p>
</body>
</html>`))
}
