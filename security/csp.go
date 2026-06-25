package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/middleware"
)

// CSPSettings defines the CSP directives.
// In a full implementation, these might be read from godjango settings.
var (
	CSPDefaultSrc     = []string{"'self'"}
	CSPScriptSrc      = []string{"'self'"}
	CSPStyleSrc       = []string{"'self'"}
	CSPImgSrc         = []string{"'self'"}
	CSPFrameAncestors = []string{"'none'"} // for clickjacking protection
)

// generateNonce creates a random base64 string for CSP nonces.
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// CSPMiddleware adds the Content-Security-Policy header to responses.
type CSPMiddleware struct{}

// NewCSPMiddleware creates a new CSPMiddleware.
func NewCSPMiddleware() *CSPMiddleware {
	return &CSPMiddleware{}
}

func (m *CSPMiddleware) Process(next middleware.Handler) middleware.Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		nonce := generateNonce()

		// Attach nonce to request context so template tags or views can access it
		req.META["CSP_NONCE"] = nonce
		req.Request = req.WithContext(godjangohttp.WithValue(req, "CSP_NONCE", nonce).Context)

		resp := next(req)

		if httpResp, ok := resp.(*godjangohttp.HttpResponse); ok {
			var policies []string

			if len(CSPDefaultSrc) > 0 {
				policies = append(policies, fmt.Sprintf("default-src %s", strings.Join(CSPDefaultSrc, " ")))
			}

			if len(CSPScriptSrc) > 0 {
				// Inject nonce into script-src
				scripts := append([]string{}, CSPScriptSrc...)
				scripts = append(scripts, fmt.Sprintf("'nonce-%s'", nonce))
				policies = append(policies, fmt.Sprintf("script-src %s", strings.Join(scripts, " ")))
			}

			if len(CSPStyleSrc) > 0 {
				policies = append(policies, fmt.Sprintf("style-src %s", strings.Join(CSPStyleSrc, " ")))
			}

			if len(CSPImgSrc) > 0 {
				policies = append(policies, fmt.Sprintf("img-src %s", strings.Join(CSPImgSrc, " ")))
			}

			if len(CSPFrameAncestors) > 0 {
				policies = append(policies, fmt.Sprintf("frame-ancestors %s", strings.Join(CSPFrameAncestors, " ")))
			}

			cspStr := strings.Join(policies, "; ")
			httpResp.Headers.Set("Content-Security-Policy", cspStr)
		}

		return resp
	}
}
