// Endpoints (all protected by Bearer token auth when ADMIN_TOKEN env var is set):

// GET  /admin/services                      — list all services with live effective config
// PATCH /admin/services/{name}/faults       — apply an in-memory fault override
// POST  /admin/services/{name}/faults/reset — delete the override, revert to YAML
// GET  /admin/config/diff                   — show drift between YAML and live overrides
package adminapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
	"github.com/tanay13/GlitchMesh/internal/shared/utils"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/admin/services", authMiddleware(http.HandlerFunc(servicesHandler)))
	mux.Handle("/admin/services/", authMiddleware(http.HandlerFunc(serviceDetailRouter)))
	mux.Handle("/admin/config/diff", authMiddleware(http.HandlerFunc(configDiffHandler)))
}

// authMiddleware enforces Bearer token authentication.
// If ADMIN_TOKEN env var is empty, auth is skipped (dev mode)
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("ADMIN_TOKEN")
		if token == "" {
			// Dev mode: no token configured, allow all
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.WriteJSONError(w, http.StatusUnauthorized, "missing or invalid Authorization header (expected: Bearer <token>)")
			return
		}

		provided := strings.TrimPrefix(authHeader, "Bearer ")
		if provided != token {
			utils.WriteJSONError(w, http.StatusUnauthorized, "invalid admin token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

type serviceListItem struct {
	Name           string       `json:"name"`
	URL            string       `json:"url"`
	EffectiveFault models.Fault `json:"effective_fault"`
	HasOverride    bool         `json:"has_override"`
}

type serviceListResponse struct {
	Generation int64             `json:"config_generation"`
	Services   []serviceListItem `json:"services"`
}

type diffEntry struct {
	ServiceName string       `json:"service"`
	BaseFault   models.Fault `json:"base_fault"`
	Override    models.Fault `json:"override"`
}

type diffResponse struct {
	Generation int64       `json:"config_generation"`
	Drifted    []diffEntry `json:"drifted_services"`
}

// GET /admin/services

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	proxy := config.GetProxyConfig()
	if proxy == nil {
		utils.WriteJSONError(w, http.StatusServiceUnavailable, "proxy config not loaded yet")
		return
	}

	overrides := config.GetServiceOverrides()
	items := make([]serviceListItem, 0, len(proxy.Service))

	for _, svc := range proxy.Service {
		effective := config.GetEffectiveServiceConfig(svc.Name)
		_, hasOverride := overrides[svc.Name]

		effectiveFault := svc.Fault
		if effective != nil {
			effectiveFault = effective.Fault
		}

		items = append(items, serviceListItem{
			Name:           svc.Name,
			URL:            svc.Url,
			EffectiveFault: effectiveFault,
			HasOverride:    hasOverride,
		})
	}

	resp := serviceListResponse{
		Generation: config.GetConfigGeneration(),
		Services:   items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Router for /admin/services/{name}/...

// serviceDetailRouter dispatches sub-paths under /admin/services/{name}/.
func serviceDetailRouter(w http.ResponseWriter, r *http.Request) {
	// Strip "/admin/services/" prefix.
	rest := strings.TrimPrefix(r.URL.Path, "/admin/services/")
	if rest == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "missing service name in path")
		return
	}

	parts := strings.SplitN(rest, "/", 3)
	serviceName := parts[0]

	if len(parts) < 2 || parts[1] != "faults" {
		utils.WriteJSONError(w, http.StatusNotFound, "unknown admin endpoint — try /admin/services/{name}/faults")
		return
	}

	// /admin/services/{name}/faults/reset
	if len(parts) == 3 && parts[2] == "reset" {
		if r.Method != http.MethodPost {
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed — use POST")
			return
		}
		faultResetHandler(w, r, serviceName)
		return
	}

	// /admin/services/{name}/faults
	if len(parts) == 2 {
		switch r.Method {
		case http.MethodPatch:
			faultPatchHandler(w, r, serviceName)
		default:
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed — use PATCH")
		}
		return
	}

	utils.WriteJSONError(w, http.StatusNotFound, "unknown admin endpoint")
}

// PATCH /admin/services/{name}/faults

func faultPatchHandler(w http.ResponseWriter, r *http.Request, serviceName string) {

	proxy := config.GetProxyConfig()
	if proxy == nil || !serviceExists(proxy, serviceName) {
		utils.WriteJSONError(w, http.StatusNotFound,
			"service '"+serviceName+"' not found in proxy configuration")
		return
	}

	var fault models.Fault
	if err := json.NewDecoder(r.Body).Decode(&fault); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if err := config.SetServiceOverride(serviceName, fault); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message":    "override applied",
		"service":    serviceName,
		"generation": config.GetConfigGeneration(),
	})
}

// POST /admin/services/{name}/faults/reset

func faultResetHandler(w http.ResponseWriter, r *http.Request, serviceName string) {
	proxy := config.GetProxyConfig()
	if proxy == nil || !serviceExists(proxy, serviceName) {
		utils.WriteJSONError(w, http.StatusNotFound,
			"service '"+serviceName+"' not found in proxy configuration")
		return
	}

	config.DeleteServiceOverride(serviceName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "override removed — service reverted to YAML config",
		"service": serviceName,
	})
}

// GET /admin/config/diff

func configDiffHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	proxy := config.GetProxyConfig()
	if proxy == nil {
		utils.WriteJSONError(w, http.StatusServiceUnavailable, "proxy config not loaded yet")
		return
	}

	overrides := config.GetServiceOverrides()
	var drifted []diffEntry

	for _, svc := range proxy.Service {
		if override, ok := overrides[svc.Name]; ok {
			drifted = append(drifted, diffEntry{
				ServiceName: svc.Name,
				BaseFault:   svc.Fault,
				Override:    override,
			})
		}
	}

	if drifted == nil {
		drifted = []diffEntry{} // return [] not null in JSON
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diffResponse{
		Generation: config.GetConfigGeneration(),
		Drifted:    drifted,
	})
}

func serviceExists(proxy *models.Proxy, name string) bool {
	for _, svc := range proxy.Service {
		if svc.Name == name {
			return true
		}
	}
	return false
}
