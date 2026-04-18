package toolbar

import (
	"fmt"
	"strings"
	"sync"
	"time"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

// TemplatesPanel shows rendered templates
type TemplatesPanel struct {
	BasePanel
	mu        sync.Mutex
	Templates []TemplateInfo
}

type TemplateInfo struct {
	Name     string
	Duration time.Duration
	Context  string
}

func NewTemplatesPanel() *TemplatesPanel { return &TemplatesPanel{} }
func (p *TemplatesPanel) Name() string { return "Templates" }
func (p *TemplatesPanel) Title() string { return "Templates" }
func (p *TemplatesPanel) NavTitle() string { return "Templates" }
func (p *TemplatesPanel) NavSubtitle() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("%d templates", len(p.Templates))
}
func (p *TemplatesPanel) RenderContent() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<h3>Templates Rendered (%d)</h3><ul>", len(p.Templates)))
	for _, t := range p.Templates {
		sb.WriteString(fmt.Sprintf("<li><strong>%s</strong> (%v) - Context keys: %s</li>", t.Name, t.Duration, t.Context))
	}
	sb.WriteString("</ul>")
	return sb.String()
}

// RecordTemplate records a template rendering event.
func RecordTemplate(req *godjangohttp.Request, name string, duration time.Duration, contextKeys []string) {
	if req == nil {
		return
	}
	tb := GetToolbarFromContext(req.Context)
	if tb == nil {
		return
	}
	if p, ok := tb.Panels["Templates"]; ok {
		if tp, ok := p.(*TemplatesPanel); ok {
			tp.mu.Lock()
			tp.Templates = append(tp.Templates, TemplateInfo{
				Name:     name,
				Duration: duration,
				Context:  strings.Join(contextKeys, ", "),
			})
			tp.mu.Unlock()
		}
	}
}

// CachePanel shows cache events
type CachePanel struct {
	BasePanel
	mu     sync.Mutex
	Events []CacheEvent
	Hits   int
	Misses int
}

type CacheEvent struct {
	Action string
	Key    string
	Hit    bool
}

func NewCachePanel() *CachePanel { return &CachePanel{} }
func (p *CachePanel) Name() string { return "Cache" }
func (p *CachePanel) Title() string { return "Cache" }
func (p *CachePanel) NavTitle() string { return "Cache" }
func (p *CachePanel) NavSubtitle() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("%d hits, %d misses", p.Hits, p.Misses)
}
func (p *CachePanel) RenderContent() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<h3>Cache (%d hits, %d misses)</h3><ul>", p.Hits, p.Misses))
	for _, e := range p.Events {
		hitStr := ""
		if e.Action == "GET" {
			if e.Hit { hitStr = " [HIT]" } else { hitStr = " [MISS]" }
		}
		sb.WriteString(fmt.Sprintf("<li>%s %s%s</li>", e.Action, e.Key, hitStr))
	}
	sb.WriteString("</ul>")
	return sb.String()
}

// LoggingPanel shows logs
type LoggingPanel struct {
	BasePanel
	mu   sync.Mutex
	Logs []string
}
func NewLoggingPanel() *LoggingPanel { return &LoggingPanel{} }
func (p *LoggingPanel) Name() string { return "Logging" }
func (p *LoggingPanel) Title() string { return "Logging" }
func (p *LoggingPanel) NavTitle() string { return "Logging" }
func (p *LoggingPanel) RenderContent() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var sb strings.Builder
	sb.WriteString("<h3>Log Records</h3><ul>")
	for _, l := range p.Logs {
		sb.WriteString(fmt.Sprintf("<li>%s</li>", l))
	}
	sb.WriteString("</ul>")
	return sb.String()
}

// SignalsPanel shows signals
type SignalsPanel struct {
	BasePanel
	mu      sync.Mutex
	Signals []string
}
func NewSignalsPanel() *SignalsPanel { return &SignalsPanel{} }
func (p *SignalsPanel) Name() string { return "Signals" }
func (p *SignalsPanel) Title() string { return "Signals" }
func (p *SignalsPanel) NavTitle() string { return "Signals" }
func (p *SignalsPanel) RenderContent() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var sb strings.Builder
	sb.WriteString("<h3>Signals Fired</h3><ul>")
	for _, s := range p.Signals {
		sb.WriteString(fmt.Sprintf("<li>%s</li>", s))
	}
	sb.WriteString("</ul>")
	return sb.String()
}

// ProfilerPanel (Optional)
type ProfilerPanel struct {
	BasePanel
}
func NewProfilerPanel() *ProfilerPanel { return &ProfilerPanel{} }
func (p *ProfilerPanel) Name() string { return "Profiler" }
func (p *ProfilerPanel) Title() string { return "Profiler" }
func (p *ProfilerPanel) NavTitle() string { return "Profiler" }
func (p *ProfilerPanel) RenderContent() string {
	return "<p>Profiler not active. Run with ?prof=1 to collect cProfile-like data.</p>"
}
