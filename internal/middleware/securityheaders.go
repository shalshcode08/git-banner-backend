package middleware

import "net/http"

// SecurityHeaders adds defensive HTTP headers to every response and handles
// CORS so the SVG images can be embedded in any origin (GitHub READMEs, etc.).
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// CORS — this is a public image API, all origins are legitimate consumers.
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, OPTIONS")

		// Preflight — browsers send OPTIONS before cross-origin GET with custom headers.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Prevent browsers from MIME-sniffing the response away from image/svg+xml.
		h.Set("X-Content-Type-Options", "nosniff")

		// Disallow embedding this API in iframes (not needed for an image API,
		// but guards against clickjacking on any HTML error pages).
		h.Set("X-Frame-Options", "DENY")

		// Disable the legacy XSS auditor (modern browsers ignore it, but harmless).
		h.Set("X-XSS-Protection", "0")

		// Only send the origin when navigating to HTTPS targets.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}
