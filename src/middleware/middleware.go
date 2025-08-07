package middleware

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Definiert ein einheitliches Zeitformat als Konstante
const timeFormat = time.RFC1123

// responseWriterWrapper ist eine Struktur, die http.ResponseWriter erweitert,
// um Statuscode und Antwortgröße mitzuprotokollieren.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// InitializeLogger konfiguriert das Logging und leitet Ausgaben sowohl in eine Datei als auch in die Konsole um.
func InitializeLogger(logFile string) error {
	// Versucht, die Log-Datei im Append-Modus zu öffnen oder zu erstellen
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// MultiWriter erstellt eine Kombination aus Datei und Konsole
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Setzt den Log-Output auf die Kombination aus Konsole und Datei
	log.SetOutput(multiWriter)

	// Optional: Setzt das Log-Format
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Zeitstempel und Dateipfad mit Zeilennummer

	return nil
}

// WriteHeader erfasst den Statuscode der Antwort.
func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write erfasst die Größe der Antwort.
func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggerMiddleware protokolliert die Details jeder eingehenden HTTP-Anfrage und -Antwort,
// einschließlich Startzeit, Methode, URL, Statuscode und Dauer der Anfrage.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Startzeit wird aufgezeichnet
		start := time.Now()
		log.Printf("Startzeit: %s | Methode: %s | URL: %s | RemoteAddr: %s",
			start.Format(timeFormat), r.Method, r.RequestURI, r.RemoteAddr)

		// Wrap den ResponseWriter, um Status und Größe zu erfassen
		wrappedWriter := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Der nächste Handler in der Kette wird aufgerufen
		next.ServeHTTP(wrappedWriter, r)

		// Endzeit und Dauer werden protokolliert
		end := time.Now()
		duration := end.Sub(start)
		log.Printf("Endzeit: %s | Dauer: %s | Statuscode: %d | Antwortgröße: %d Bytes",
			end.Format(timeFormat), duration, wrappedWriter.statusCode, wrappedWriter.size)
	})
}

// NoCacheMiddleware verhindert das Caching von HTTP-Antworten, um sicherzustellen,
// dass der Client immer die aktuellsten Daten erhält.
func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Setzt mehrere Header, um sicherzustellen, dass der Client die Antwort nicht cached.
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Der nächste Handler in der Kette wird aufgerufen
		next.ServeHTTP(w, r)
	})
}
