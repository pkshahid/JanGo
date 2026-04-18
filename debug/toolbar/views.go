package toolbar

import (
	"net/http"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/urls"
)

// RenderPanel is the AJAX endpoint view that renders a specific panel's content.
func RenderPanel(req *godjangohttp.Request) godjangohttp.Response {
	req.ParseForm()
	storeID := req.GET.Get("store_id")
	panelID := req.GET.Get("panel_id")

	if storeID == "" || panelID == "" {
		return godjangohttp.HttpResponseBadRequest("Missing parameters")
	}

	tb := GetToolbar(storeID)
	if tb == nil {
		return godjangohttp.HttpResponseNotFound("Toolbar instance not found")
	}

	panel, ok := tb.Panels[panelID]
	if !ok {
		return godjangohttp.HttpResponseNotFound("Panel not found")
	}

	content := panel.RenderContent()

	resp := godjangohttp.NewHttpResponse(content, http.StatusOK)
	// Optionally set no-cache headers here
	return resp
}

// RegisterRoutes registers the debug toolbar routes with the given router.
func RegisterRoutes(router *urls.Router) {
	// The prefix used in middleware injection is /djdt/
	// So we register /djdt/render_panel/
	router.Add(urls.Path("/djdt/render_panel/", RenderPanel, "djdt_render_panel", nil))
}
