package toolbar

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type SQLPanel struct {
	BasePanel
	mu         sync.Mutex
	Queries    []SQLQueryInfo
	TotalTime  time.Duration
	Duplicates int
}

func NewSQLPanel() *SQLPanel {
	return &SQLPanel{}
}

func (p *SQLPanel) Name() string     { return "SQL" }
func (p *SQLPanel) Title() string    { return "SQL Queries" }
func (p *SQLPanel) NavTitle() string { return "SQL" }
func (p *SQLPanel) NavSubtitle() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("%d queries, %v", len(p.Queries), p.TotalTime)
}

func (p *SQLPanel) RenderContent() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<h3>SQL Queries (%d total, %d duplicates, %v)</h3>", len(p.Queries), p.Duplicates, p.TotalTime))
	sb.WriteString("<table border='1' style='width:100%; text-align:left;'>")
	sb.WriteString("<tr><th>Time</th><th>Query</th><th>Params</th><th>Stack</th></tr>")

	for _, q := range p.Queries {
		style := ""
		if q.Duplicate {
			style = "background-color: #fdd;"
		}
		sb.WriteString(fmt.Sprintf("<tr style='%s'>", style))
		sb.WriteString(fmt.Sprintf("<td>%v</td>", q.Duration))
		sb.WriteString(fmt.Sprintf("<td><pre>%s</pre></td>", q.Query))
		sb.WriteString(fmt.Sprintf("<td>%v</td>", q.Params))
		sb.WriteString(fmt.Sprintf("<td><pre style='font-size:10px; max-height:100px; overflow-y:auto;'>%s</pre></td>", q.Traceback))
		sb.WriteString("</tr>")
	}

	sb.WriteString("</table>")
	return sb.String()
}
