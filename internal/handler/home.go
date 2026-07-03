package handler

import (
	"net/http"
)

// HomeHandler responds at the API root. The storefront UI is served by the
// separate Next.js frontend, so this just returns a small JSON pointer.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "sby-twilight-api",
		"status":  "ok",
		"docs":    "/api",
	})
}
