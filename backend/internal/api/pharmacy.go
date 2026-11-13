package api

import (
	"net/http"

	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
)

func (s *Server) pharmacyCatalog(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	s.catalog(w, r)
}

func (s *Server) emptyPharmacyResource(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": []map[string]any{}})
}
