package toolbar

import (
	godjangohttp "github.com/pkshahid/JanGo/http"
)

// Panel defines the interface for debug toolbar panels.
type Panel interface {
	Name() string
	Title() string
	Enable() bool
	ProcessRequest(req *godjangohttp.Request)
	ProcessResponse(req *godjangohttp.Request, resp godjangohttp.Response)
	RenderContent() string
	HasContent() bool
	NavTitle() string
	NavSubtitle() string
}

// BasePanel provides default implementations for Panel.
type BasePanel struct{}

func (b *BasePanel) Enable() bool { return true }
func (b *BasePanel) ProcessRequest(req *godjangohttp.Request) {}
func (b *BasePanel) ProcessResponse(req *godjangohttp.Request, resp godjangohttp.Response) {}
func (b *BasePanel) RenderContent() string { return "" }
func (b *BasePanel) HasContent() bool { return true }
func (b *BasePanel) NavTitle() string { return "" }
func (b *BasePanel) NavSubtitle() string { return "" }
