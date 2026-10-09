// Package web embeds the panel's static files into the server binary.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var files embed.FS

// policy allows only same-origin resources: no inline scripts or styles, no
// eval and no third-party hosts. Alpine is served locally in its CSP build.
const policy = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// Handler serves index.html at / and the assets under /static/.
func Handler() http.Handler {
	static, e := fs.Sub(files, "static")
	if e != nil {
		panic(e)
	}
	assets := http.StripPrefix("/static/", http.FileServer(http.FS(static)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", policy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Cache-Control", "no-cache")
		if r.URL.Path == "/" {
			http.ServeFileFS(w, r, static, "index.html")
			return
		}
		assets.ServeHTTP(w, r)
	})
}
