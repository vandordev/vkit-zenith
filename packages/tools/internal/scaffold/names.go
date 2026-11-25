package scaffold

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	handlerName  = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)+$`)
	snakeName    = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	domainName   = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
	exportedName = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	pathParam    = regexp.MustCompile(`\{([^}]*)\}`)
	paramName    = regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`)
)

// apps maps each scaffoldable API to its operation-ID prefix; the public API uses bare camelCase (ADR-0007).
var apps = map[string]string{"web-api": "webV1", "bo-api": "boV1", "api": ""}

var methods = map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}

// reserved application packages a usecase scaffold must never write into.
var reserved = map[string]bool{"internal": true, "test": true}

func pascal(snake string) string {
	var b strings.Builder
	for _, part := range strings.Split(snake, "_") {
		if part != "" {
			b.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return b.String()
}

// split turns resource-first file names into resource and action: projects_archive -> projects, archive.
func split(name string) (resource, action string) {
	i := strings.LastIndex(name, "_")
	return name[:i], name[i+1:]
}

func (o HandlerOptions) validate() ([]string, error) {
	if _, ok := apps[o.App]; !ok {
		return nil, fmt.Errorf("app %q: want web-api, bo-api or api", o.App)
	}
	if !handlerName.MatchString(o.Name) {
		return nil, fmt.Errorf("name %q: want snake_case resource_action, e.g. projects_list", o.Name)
	}
	if o.Kind != "query" && o.Kind != "mutation" {
		return nil, fmt.Errorf("kind %q: want query or mutation", o.Kind)
	}
	if !methods[o.Method] {
		return nil, fmt.Errorf("method %q: want GET, POST, PUT, PATCH or DELETE", o.Method)
	}
	if o.Kind == "query" && o.Method != "GET" {
		return nil, fmt.Errorf("query handlers use GET, got %s", o.Method)
	}
	if o.Kind == "mutation" && o.Method == "GET" {
		return nil, fmt.Errorf("mutation handlers never use GET")
	}
	if !strings.HasPrefix(o.Path, "/v1/") || strings.ContainsAny(o.Path, " ?#") {
		return nil, fmt.Errorf("path %q: want a versioned path starting with /v1/", o.Path)
	}
	var params []string
	for _, m := range pathParam.FindAllStringSubmatch(o.Path, -1) {
		if !paramName.MatchString(m[1]) {
			return nil, fmt.Errorf("path parameter %q: want lowerCamel, e.g. {projectId}", m[1])
		}
		params = append(params, m[1])
	}
	if o.App == "web-api" && !strings.Contains(o.Path, "{organizationId}") {
		return nil, fmt.Errorf("web-api paths are scoped by {organizationId} (ADR-0002)")
	}
	return params, nil
}

func (o UsecaseOptions) validate() error {
	if !domainName.MatchString(o.Domain) || reserved[o.Domain] {
		return fmt.Errorf("domain %q: want a lowercase package name such as projects", o.Domain)
	}
	if !exportedName.MatchString(o.Name) {
		return fmt.Errorf("usecase %q: want an exported verb such as Create", o.Name)
	}
	return nil
}
