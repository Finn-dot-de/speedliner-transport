package middleware

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"log/slog"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

const (
	timeFormat    = time.RFC3339Nano
	reqIDHeader   = "X-Request-ID"
	slowThreshold = 1 * time.Second
)

// ---- Konfiguration (ENV-Overrides in InitializeLogger) ----
var (
	maxUARefLen     = 120  // UA/Referrer hart kürzen
	assetSkipFast   = true // schnelle Asset-Requests (< slowThreshold, <400) komplett skippen
	assetSampleN    = 0    // 0 = kein Sampling; >0 = jeden N-ten Asset-Request loggen (falls nicht geskippt)
	assetExts       = []string{".css", ".js", ".map", ".png", ".jpg", ".jpeg", ".webp", ".ico", ".svg", ".gif", ".woff", ".woff2", ".ttf"}
	loggerLevel     = slog.LevelInfo // Default-Level
	initializedSlog atomic.Bool
	assetCounter    atomic.Uint64
	twoXXSampleN    = 0 // optionales Sampling für unauffällige 2xx (0=aus)
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	wroteHeader bool
}

// InitializeLogger konfiguriert Stdlog + slog (JSON) auf Konsole + Datei.
// ENV-Variablen (optional):
//
//	LOG_LEVEL=debug|info|warn|error
//	LOG_ASSET_SKIP_FAST=true|false
//	LOG_ASSET_SAMPLE_N=10           (jeden 10ten Asset-Request)
//	LOG_2XX_SAMPLE_N=0              (Sampling für „langweilige“ 2xx)
//	LOG_UA_REF_MAXLEN=120
func InitializeLogger(logFile string) error {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	multi := io.MultiWriter(os.Stdout, file)

	// stdlib log → multi
	log.SetOutput(multi)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ENV einlesen
	if v := strings.ToLower(os.Getenv("LOG_LEVEL")); v != "" {
		switch v {
		case "debug":
			loggerLevel = slog.LevelDebug
		case "info":
			loggerLevel = slog.LevelInfo
		case "warn", "warning":
			loggerLevel = slog.LevelWarn
		case "error":
			loggerLevel = slog.LevelError
		}
	}
	if v := os.Getenv("LOG_ASSET_SKIP_FAST"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			assetSkipFast = b
		}
	}
	if v := os.Getenv("LOG_ASSET_SAMPLE_N"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			assetSampleN = n
		}
	}
	if v := os.Getenv("LOG_2XX_SAMPLE_N"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			twoXXSampleN = n
		}
	}
	if v := os.Getenv("LOG_UA_REF_MAXLEN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxUARefLen = n
		}
	}

	// slog → JSON auf multi
	h := slog.NewJSONHandler(multi, &slog.HandlerOptions{Level: loggerLevel})
	slog.SetDefault(slog.New(h))
	initializedSlog.Store(true)

	// deterministische, aber unterschiedliche Seeds (für Sampling)

	return nil
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	if !lw.wroteHeader {
		lw.status = code
		lw.wroteHeader = true
		lw.ResponseWriter.WriteHeader(code)
	}
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	if !lw.wroteHeader {
		lw.WriteHeader(http.StatusOK)
	}
	n, err := lw.ResponseWriter.Write(b)
	lw.size += n
	return n, err
}

// Interface passthroughs
func (lw *loggingResponseWriter) Flush() {
	if f, ok := lw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
func (lw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := lw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacker not supported")
	}
	return h.Hijack()
}
func (lw *loggingResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := lw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

// LoggerMiddleware protokolliert strukturiert + robust.
// Features:
//   - Request-ID
//   - Debug-Start-Log
//   - Asset-Filter/Sampling
//   - Level nach Status/Latenz
//   - Panic-Recovery
type contextKey string

func LoggerMiddleware(next http.Handler) http.Handler {
	logger := slog.Default()
	// Fallback, falls InitializeLogger nicht aufgerufen wurde
	if !initializedSlog.Load() {
		h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel})
		logger = slog.New(h)
		slog.SetDefault(logger)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := ensureRequestID(r)
		w.Header().Set(reqIDHeader, reqID)

		clientIP := clientIPFromRequest(r)
		ua := truncate(r.UserAgent(), maxUARefLen)
		ref := truncate(r.Referer(), maxUARefLen)
		cl := r.Header.Get("Content-Length")

		ctx := context.WithValue(r.Context(), contextKey(reqIDHeader), reqID)
		r = r.WithContext(ctx)
		log := logger.With(
			"req.id", reqID,
			"req.method", r.Method,
			"req.path", r.URL.Path,
			"req.query", r.URL.RawQuery,
			"req.ip", clientIP,
			"req.ua", ua,
			"req.ref", ref,
			"req.content_length", cl,
		)

		// Start-Log (Debug)
		log.Debug("request started", "ts", start.Format(timeFormat))

		// Response wrappen
		lw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}

		// Panic-Recovery + Abschlusslog
		defer func() {
			if rec := recover(); rec != nil {
				lw.WriteHeader(http.StatusInternalServerError)
				log.Error("panic recovered",
					"error", rec,
					"duration_ms", durationMs(time.Since(start)),
				)
			}

			d := time.Since(start)
			status := lw.status
			isAsset := isAssetPath(r.URL.Path)

			// „Langweilige“ Assets ggf. komplett überspringen
			if isAsset && status < 400 && d < slowThreshold && assetSkipFast {
				// mit Sampling optional trotzdem ab und zu loggen
				if assetSampleN > 0 {
					if !everyNth(assetSampleN, &assetCounter) {
						return
					}
				} else {
					return
				}
			}

			// Optional: Sampling für unauffällige 2xx (nicht-Asset)
			if !isAsset && status >= 200 && status < 300 && d < slowThreshold && twoXXSampleN > 0 {
				// kleine zufällige Streuung, damit nicht streng periodisch
				if rand.Intn(twoXXSampleN) != 0 {
					return
				}
			}

			// Level bestimmen
			level := levelFor(status, d)
			log.Log(context.Background(), level, "request finished",
				"res.status", status,
				"res.bytes", lw.size,
				"duration_ms", durationMs(d),
			)
		}()

		next.ServeHTTP(lw, r)
	})
}

// ---- helpers ----

func ensureRequestID(r *http.Request) string {
	if id := r.Header.Get(reqIDHeader); id != "" {
		return id
	}
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func clientIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return rip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func sanitizeHeaders(h http.Header) http.Header {
	const maxVals = 3
	clone := http.Header{}
	secret := map[string]struct{}{
		"Authorization": {}, "Cookie": {}, "Set-Cookie": {},
		"Proxy-Authorization": {}, "X-Api-Key": {},
	}
	for k, v := range h {
		if _, ok := secret[k]; ok {
			clone[k] = []string{"<redacted>"}
			continue
		}
		if len(v) > maxVals {
			clone[k] = append([]string{}, v[:maxVals]...)
			clone[k] = append(clone[k], "...(truncated)")
		} else {
			clone[k] = append([]string{}, v...)
		}
	}
	return clone
}

// NoCacheMiddleware verhindert Caching.
func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

type visitor struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

var (
	visitors sync.Map
	ttl      = 10 * time.Minute // Besucher-Einträge nach Inaktivität aufräumen
	rps      = rate.Limit(5)    // 5 Requests pro Sekunde
	burst    = 10               // bis zu 10 auf einmal erlauben
	ticker   = time.NewTicker(5 * time.Minute)
)

func init() {
	// Periodisches Cleanup
	go func() {
		for range ticker.C {
			now := time.Now()
			visitors.Range(func(k, v any) bool {
				if now.Sub(v.(*visitor).lastSeen) > ttl {
					visitors.Delete(k)
				}
				return true
			})
		}
	}()
}

func keyFor(r *http.Request) string {
	ip := clientIPFromRequest(r)
	// ggf. Pfade grob gruppieren, damit Assets nicht pro Datei zählen:
	path := r.URL.Path
	switch {
	case strings.HasPrefix(path, "/assets/"):
		path = "/assets/*"
	case strings.HasPrefix(path, "/app/login"):
		path = "/app/login"
	}
	return ip + "|" + path
}

func getLimiter(key string) *rate.Limiter {
	now := time.Now()
	if v, ok := visitors.Load(key); ok {
		vis := v.(*visitor)
		vis.lastSeen = now
		return vis.lim
	}
	lim := rate.NewLimiter(rps, burst)
	visitors.Store(key, &visitor{lim: lim, lastSeen: now})
	return lim
}

// RateLimit: sanftes Ratenlimit mit Burst; optional nur auf „write“-Methoden anwenden.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Optional: nur auf sensible Methoden/Pfade limitieren
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		lim := getLimiter(keyFor(r))
		if !lim.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---- kleine Utils ----

func isAssetPath(p string) bool {
	ext := strings.ToLower(path.Ext(p))
	if ext == "" {
		return false
	}
	for _, e := range assetExts {
		if ext == e {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	// Unicode-sicher kürzen ist aufwändiger; hier einfacher Cut + Suffix
	const suffix = "…"
	if n <= len(suffix) {
		return s[:n]
	}
	return s[:n-len(suffix)] + suffix
}

func levelFor(status int, d time.Duration) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	case d >= slowThreshold:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func durationMs(d time.Duration) int64 {
	// konsistente Millisekunden (aufrunden für 0.x ms)
	return int64(math.Round(float64(d) / float64(time.Millisecond)))
}

// everyNth: atomar jeden N-ten Treffer true
func everyNth(n int, c *atomic.Uint64) bool {
	if n <= 1 {
		return true
	}
	v := c.Add(1)
	return v%uint64(n) == 0
}
