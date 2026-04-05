package urls

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	godjangohttp "github.com/godjango/godjango/http"
)

// ViewFunc represents a view handling a request.
type ViewFunc func(req *godjangohttp.Request) godjangohttp.Response

// ResolverMatch contains information about the matched URL.
type ResolverMatch struct {
	Func      ViewFunc
	Args      []string
	Kwargs    map[string]any
	URLName   string
	AppName   string
	Namespace string
}

// URLPattern represents a single route.
type URLPattern struct {
	Pattern      string
	View         ViewFunc
	Name         string
	Extra        map[string]any
	IsPrefix     bool
	Include      *URLconf
	RegexPattern *regexp.Regexp
	Converters   map[string]string // param -> converter
}

// URLconf represents a list of URL patterns.
type URLconf struct {
	Patterns  []*URLPattern
	AppName   string
	Namespace string
}

var (
	routePatternRegex = regexp.MustCompile(`<([^>]+)>`)
	converters        = map[string]string{
		"int":  `[0-9]+`,
		"slug": `[-a-zA-Z0-9_]+`,
		"uuid": `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		"str":  `[^/]+`,
	}
)

// parseRoutePattern converts Django-style paths like <int:year> into compiled regexes.
func parseRoutePattern(pattern string, isPrefix bool) (*regexp.Regexp, map[string]string) {
	convMap := make(map[string]string)

	regexStr := routePatternRegex.ReplaceAllStringFunc(pattern, func(match string) string {
		inner := match[1 : len(match)-1]
		parts := strings.SplitN(inner, ":", 2)

		var conv, name string
		if len(parts) == 2 {
			conv = parts[0]
			name = parts[1]
		} else {
			conv = "str"
			name = parts[0]
		}

		convMap[name] = conv
		regexPart, ok := converters[conv]
		if !ok {
			regexPart = `[^/]+`
		}
		return fmt.Sprintf("(?P<%s>%s)", name, regexPart)
	})

	if !isPrefix {
		regexStr = "^" + regexStr + "$"
	} else {
		regexStr = "^" + regexStr
	}
	return regexp.MustCompile(regexStr), convMap
}

// Path creates a standard URLPattern.
func Path(pattern string, view ViewFunc, name string, extra map[string]any) *URLPattern {
	re, convMap := parseRoutePattern(pattern, false)
	return &URLPattern{
		Pattern:      pattern,
		View:         view,
		Name:         name,
		Extra:        extra,
		RegexPattern: re,
		Converters:   convMap,
	}
}

// RePath creates a regex-based URLPattern.
func RePath(pattern string, view ViewFunc, name string, extra map[string]any) *URLPattern {
	return &URLPattern{
		Pattern:      pattern,
		View:         view,
		Name:         name,
		Extra:        extra,
		RegexPattern: regexp.MustCompile("^" + pattern + "$"),
	}
}

// Include creates a prefix URLPattern to include another URLconf.
func Include(prefix string, urlconf *URLconf) *URLPattern {
	re, convMap := parseRoutePattern(prefix, true)
	return &URLPattern{
		Pattern:      prefix,
		IsPrefix:     true,
		Include:      urlconf,
		RegexPattern: re,
		Converters:   convMap,
	}
}

// Router handles pattern matching and reversing.
type Router struct {
	mu       sync.RWMutex
	patterns []*URLPattern
}

var globalRouter = &Router{}

// GetGlobalRouter returns the global router instance.
func GetGlobalRouter() *Router {
	return globalRouter
}

// Add adds a pattern to the router.
func (r *Router) Add(pattern *URLPattern) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.patterns = append(r.patterns, pattern)
}

// Match attempts to find a matching URLPattern for the given path.
func (r *Router) Match(path string) (*ResolverMatch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.matchPatterns(path, r.patterns, "", "")
}

func (r *Router) matchPatterns(path string, patterns []*URLPattern, currentNamespace, currentAppName string) (*ResolverMatch, error) {
	for _, p := range patterns {
		matches := p.RegexPattern.FindStringSubmatch(path)
		if matches != nil {
			kwargs := make(map[string]any)
			// Initialize with extra kwargs
			for ek, ev := range p.Extra {
				kwargs[ek] = ev
			}

			// Extract named groups
			for i, name := range p.RegexPattern.SubexpNames() {
				if i != 0 && name != "" {
					kwargs[name] = matches[i]
				}
			}

			if p.IsPrefix {
				remainingPath := path[len(matches[0]):]
				// If we match a prefix but there is no leading slash, and remaining does not start with one, add it for inner match if needed.
				// For Django compatibility, usually prefix ends with slash or inner starts with slash.

				newNamespace := currentNamespace
				if p.Include.Namespace != "" {
					if newNamespace != "" {
						newNamespace += ":" + p.Include.Namespace
					} else {
						newNamespace = p.Include.Namespace
					}
				}

				newAppName := currentAppName
				if p.Include.AppName != "" {
					newAppName = p.Include.AppName
				}

				match, err := r.matchPatterns(remainingPath, p.Include.Patterns, newNamespace, newAppName)
				if err == nil {
					// Merge kwargs
					for k, v := range kwargs {
						if _, ok := match.Kwargs[k]; !ok {
							match.Kwargs[k] = v
						}
					}
					return match, nil
				}
			} else {
				fullName := p.Name
				if currentNamespace != "" && p.Name != "" {
					fullName = currentNamespace + ":" + p.Name
				}

				return &ResolverMatch{
					Func:      p.View,
					Kwargs:    kwargs,
					URLName:   fullName,
					AppName:   currentAppName,
					Namespace: currentNamespace,
				}, nil
			}
		}
	}
	return nil, errors.New("no match found")
}

// reversePatterns recursively finds the pattern and builds the URL.
func reversePatterns(name string, kwargs map[string]any, patterns []*URLPattern, currentNamespace string) (string, error) {
	for _, p := range patterns {
		if p.IsPrefix {
			newNamespace := currentNamespace
			if p.Include.Namespace != "" {
				if newNamespace != "" {
					newNamespace += ":" + p.Include.Namespace
				} else {
					newNamespace = p.Include.Namespace
				}
			}

			// We only want to search inside if the name matches the namespace prefix or if we are searching globally.
			// For simplicity, we search all includes.
			if suffix, err := reversePatterns(name, kwargs, p.Include.Patterns, newNamespace); err == nil {
				// Reconstruct prefix
				prefix := p.Pattern

				// Very basic reverse for prefix variables
				for k, v := range kwargs {
					strVal := fmt.Sprintf("%v", v)
					prefix = regexp.MustCompile(fmt.Sprintf(`<%s:%s>|<%s>`, `[^>]+`, k, k)).ReplaceAllString(prefix, strVal)
				}

				return prefix + suffix, nil
			}
		} else {
			fullName := p.Name
			if currentNamespace != "" && p.Name != "" {
				fullName = currentNamespace + ":" + p.Name
			}

			if fullName == name {
				urlStr := p.Pattern
				// Replace kwargs
				for k, v := range kwargs {
					strVal := fmt.Sprintf("%v", v)
					urlStr = regexp.MustCompile(fmt.Sprintf(`<%s:%s>|<%s>`, `[^>]+`, k, k)).ReplaceAllString(urlStr, strVal)
				}

				// Check if any unbound parameters remain
				if routePatternRegex.MatchString(urlStr) {
					return "", fmt.Errorf("missing keyword arguments for %s", name)
				}

				return urlStr, nil
			}
		}
	}
	return "", errors.New("not found")
}

// Reverse reverses a URL by name.
func (r *Router) Reverse(name string, kwargs map[string]any) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return reversePatterns(name, kwargs, r.patterns, "")
}

// Global helper
func Reverse(name string, kwargs map[string]any) string {
	urlStr, err := globalRouter.Reverse(name, kwargs)
	if err != nil {
		return ""
	}
	return "/" + urlStr // Return relative path
}
