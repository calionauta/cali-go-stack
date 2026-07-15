// SCOPE:core - /api discovery handler.
//
// Why this file exists:
// The PocketBase startup banner advertises "REST API:
// http://...:8080/api/". Without these exact routes, Go 1.22
// ServeMux subtree matching routes /api to our "/" handler (the
// Todo index). The user saw the full Todo page when they typed
// /api in their browser — wrong response for a /api URL.
//
// What PB itself registers: apis.NewRouter() (called by pb.Start)
// mounts sub-routes via Group("/api") — /api/health,
// /api/collections/..., /api/records/..., /api/realtime, etc. PB
// does NOT register an exact GET /api or GET /api/, so the ServeMux
// fallback runs.
//
// What we do here: register exact /api AND /api/ that return a
// short JSON envelope redirecting callers to PB's REST docs and to
// our top-level docs. No endpoint list — that would drift from the
// real routing once anyone adds a feature. ~20 lines, no templ,
// no separate HTML page.
package router

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// apiIndex responds to GET /api and GET /api/. Always JSON, since
// both browsers and scripts expect a clear error from a "list-endpoints"
// URL anyway. Trailing-slash variants behave identically.
func apiIndex(c *core.RequestEvent) error {
	env := map[string]any{
		"message":            "REST API root. Sub-routes are mounted by PocketBase and the app handlers under /api",
		"see":                "https://pocketbase.io/docs/api/",
		"health":             "/api/health",
		"realtime":           "/api/realtime",
		"superuserDashboard": "/_/",
	}
	return c.JSON(http.StatusOK, env)
}
