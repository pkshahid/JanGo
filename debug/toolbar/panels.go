package toolbar

import (
	"fmt"
	"strings"
	"time"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

// TimerPanel measures total request time
type TimerPanel struct {
	BasePanel
	Start time.Time
	End   time.Time
}

func NewTimerPanel() *TimerPanel { return &TimerPanel{} }
func (p *TimerPanel) Name() string { return "Timer" }
func (p *TimerPanel) Title() string { return "Time" }
func (p *TimerPanel) ProcessRequest(req *godjangohttp.Request) { p.Start = time.Now() }
func (p *TimerPanel) ProcessResponse(req *godjangohttp.Request, resp godjangohttp.Response) { p.End = time.Now() }
func (p *TimerPanel) NavTitle() string { return "Time" }
func (p *TimerPanel) NavSubtitle() string { return fmt.Sprintf("%v", p.End.Sub(p.Start)) }
func (p *TimerPanel) RenderContent() string {
	return fmt.Sprintf("<h3>Request Time</h3><p>Total Time: %v</p>", p.End.Sub(p.Start))
}

// RequestPanel shows request details
type RequestPanel struct {
	BasePanel
	Req *godjangohttp.Request
}

func NewRequestPanel() *RequestPanel { return &RequestPanel{} }
func (p *RequestPanel) Name() string { return "Request" }
func (p *RequestPanel) Title() string { return "Request" }
func (p *RequestPanel) ProcessRequest(req *godjangohttp.Request) { p.Req = req }
func (p *RequestPanel) NavTitle() string { return "Request" }
func (p *RequestPanel) RenderContent() string {
	if p.Req == nil { return "" }
	var sb strings.Builder
	sb.WriteString("<h3>GET Parameters</h3><ul>")
	for k, v := range p.Req.GET { sb.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v)) }
	sb.WriteString("</ul><h3>POST Parameters</h3><ul>")
	if p.Req.POST != nil {
		for k, v := range p.Req.POST { sb.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v)) }
	}
	sb.WriteString("</ul><h3>Cookies</h3><ul>")
	for k, v := range p.Req.COOKIES { sb.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v)) }
	sb.WriteString("</ul>")
	return sb.String()
}

// HeadersPanel shows HTTP headers
type HeadersPanel struct {
	BasePanel
	ReqHeaders  map[string][]string
	RespHeaders map[string][]string
}

func NewHeadersPanel() *HeadersPanel { return &HeadersPanel{} }
func (p *HeadersPanel) Name() string { return "Headers" }
func (p *HeadersPanel) Title() string { return "Headers" }
func (p *HeadersPanel) ProcessRequest(req *godjangohttp.Request) { p.ReqHeaders = req.Header }
func (p *HeadersPanel) ProcessResponse(req *godjangohttp.Request, resp godjangohttp.Response) {
	// Can't easily get headers from Response interface generically without casting.
	// As a workaround for this task, we will just show request headers.
}
func (p *HeadersPanel) NavTitle() string { return "Headers" }
func (p *HeadersPanel) RenderContent() string {
	var sb strings.Builder
	sb.WriteString("<h3>Request Headers</h3><ul>")
	for k, v := range p.ReqHeaders { sb.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v)) }
	sb.WriteString("</ul>")
	return sb.String()
}

// RoutingPanel shows URL matching info
type RoutingPanel struct {
	BasePanel
	Req *godjangohttp.Request
}

func NewRoutingPanel() *RoutingPanel { return &RoutingPanel{} }
func (p *RoutingPanel) Name() string { return "Routing" }
func (p *RoutingPanel) Title() string { return "Routing" }
func (p *RoutingPanel) ProcessRequest(req *godjangohttp.Request) { p.Req = req }
func (p *RoutingPanel) NavTitle() string { return "Routing" }
func (p *RoutingPanel) RenderContent() string {
	if p.Req == nil || p.Req.ResolverMatch == nil { return "<p>No routing match found.</p>" }
	var sb strings.Builder
	sb.WriteString("<h3>Routing Information</h3>")
	// ResolverMatch is generic any in Request to avoid circular import, we format it as string
	sb.WriteString(fmt.Sprintf("<p>Match: %+v</p>", p.Req.ResolverMatch))
	return sb.String()
}
