package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Langa2007/POS/backend/internal/core/customfields"
	"github.com/Langa2007/POS/backend/internal/core/importexport"
	"github.com/Langa2007/POS/backend/internal/core/rules"
	"github.com/Langa2007/POS/backend/internal/core/templates"
	"github.com/Langa2007/POS/backend/internal/core/workflows"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) listPresets(w http.ResponseWriter, r *http.Request) {
	list, err := s.presets.List(r.Context())
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"presets": list})
}

func (s *Server) applyPreset(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	presetID := r.PathValue("presetID")
	if presetID == "" {
		writeError(w, 400, "request.invalid", "preset_id is required")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return s.presets.Apply(r.Context(), tx, scope.TenantID, presetID)
	})
	if err != nil {
		writeError(w, 400, "request.invalid", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"applied": presetID})
}

func (s *Server) listCustomFields(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	entity := r.URL.Query().Get("entity_type")
	var list []customfields.Def
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		list, e = s.customfields.ListTx(r.Context(), tx, scope.TenantID, entity)
		return e
	})
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"custom_fields": list})
}

func (s *Server) upsertCustomField(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var d customfields.Def
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var out customfields.Def
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		out, e = s.customfields.Upsert(r.Context(), tx, scope.TenantID, d)
		return e
	})
	if err != nil {
		writeError(w, 400, "request.invalid", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var list []rules.Rule
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		list, e = s.rules.ListTx(r.Context(), tx, scope.TenantID, false)
		return e
	})
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"rules": list})
}

func (s *Server) upsertRule(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var ru rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&ru); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var out rules.Rule
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		out, e = s.rules.Upsert(r.Context(), tx, scope.TenantID, ru)
		return e
	})
	if err != nil {
		writeError(w, 400, "request.invalid", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) simulateRules(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var req rules.EvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var matched []map[string]any
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		matched, e = s.rules.SimulateTx(r.Context(), tx, scope.TenantID, req, time.Now().UTC())
		return e
	})
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"matched": matched})
}

func (s *Server) listWorkflows(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	out := make([]workflows.Workflow, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT id,key,name,definition,enabled FROM workflows WHERE tenant_id=$1 ORDER BY key`, scope.TenantID)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var wf workflows.Workflow
			if e = rows.Scan(&wf.ID, &wf.Key, &wf.Name, &wf.Definition, &wf.Enabled); e != nil {
				return e
			}
			out = append(out, wf)
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"workflows": out})
}

func (s *Server) upsertWorkflow(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var wf workflows.Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var out workflows.Workflow
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		out, e = s.workflows.Upsert(r.Context(), tx, scope.TenantID, wf)
		return e
	})
	if err != nil {
		writeError(w, 400, "request.invalid", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) listTemplates(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	out := make([]templates.Template, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT id,key,name,type,format,content,branch_id,version,active FROM document_templates WHERE tenant_id=$1 ORDER BY key`, scope.TenantID)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var t templates.Template
			var bid *string
			if e = rows.Scan(&t.ID, &t.Key, &t.Name, &t.Type, &t.Format, &t.Content, &bid, &t.Version, &t.Active); e != nil {
				return e
			}
			t.BranchID = bid
			out = append(out, t)
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"templates": out})
}

func (s *Server) upsertTemplate(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var t templates.Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var out templates.Template
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		out, e = s.templates.Upsert(r.Context(), tx, scope.TenantID, t)
		return e
	})
	if err != nil {
		writeError(w, 400, "request.invalid", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) renderTemplate(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	key := r.PathValue("key")
	var data any
	_ = json.NewDecoder(r.Body).Decode(&data)
	if data == nil {
		data = map[string]any{}
	}
	var out string
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		out, e = s.templates.RenderTx(r.Context(), tx, scope.TenantID, key, data)
		return e
	})
	if err != nil {
		writeError(w, 404, "template.not_found", "Template not found")
		return
	}
	writeJSON(w, 200, map[string]string{"rendered": out})
}

func (s *Server) createImportJob(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var req struct {
		EntityType string          `json:"entity_type"`
		Filename   string          `json:"filename"`
		Mapping    json.RawMessage `json:"mapping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var out importexport.Job
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		out, e = s.importexport.Create(r.Context(), tx, scope.TenantID, req.EntityType, req.Filename, req.Mapping)
		return e
	})
	if err != nil {
		writeError(w, 400, "request.invalid", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) dryRunImport(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	jobID := r.PathValue("jobID")
	var req struct {
		Rows [][]string `json:"rows"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "request.invalid", "Invalid payload")
		return
	}
	var errs []string
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var e error
		errs, e = s.importexport.DryRun(r.Context(), tx, jobID, req.Rows)
		return e
	})
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"errors": errs})
}

func (s *Server) listCapabilities(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Pool().Query(r.Context(), `SELECT id,name,description,requires,conflicts,settings FROM capability_definitions ORDER BY id`)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, desc string
		var requires, conflicts, settings []byte
		if err := rows.Scan(&id, &name, &desc, &requires, &conflicts, &settings); err != nil {
			writeError(w, 500, "internal.error", "Request could not be completed")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "description": desc, "requires": json.RawMessage(requires), "conflicts": json.RawMessage(conflicts), "settings": json.RawMessage(settings)})
	}
	writeJSON(w, 200, map[string]any{"capabilities": out})
}
