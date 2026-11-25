package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templateName = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func interpolate(value string, environment map[string]string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(value); {
		if strings.HasPrefix(value[i:], "$${") {
			out.WriteString("${")
			i += 3
			continue
		}
		if !strings.HasPrefix(value[i:], "${") {
			out.WriteByte(value[i])
			i++
			continue
		}
		end := strings.IndexByte(value[i+2:], '}')
		if end < 0 {
			return "", fmt.Errorf("malformed template")
		}
		end += i + 2
		body := value[i+2 : end]
		name, fallback, hasFallback := strings.Cut(body, ":-")
		if !templateName.MatchString(name) || strings.ContainsAny(body, "{}") {
			return "", fmt.Errorf("malformed template")
		}
		if environment != nil {
			resolved := environment[name]
			if resolved != "" {
				out.WriteString(resolved)
			} else if hasFallback {
				out.WriteString(fallback)
			} else {
				return "", fmt.Errorf("missing environment variable %s", name)
			}
		}
		i = end + 1
	}
	return out.String(), nil
}
func (d *document) valueAt(path ...string) (any, error) {
	var value any = d.root
	for _, key := range path {
		switch current := value.(type) {
		case map[string]any:
			value = current[key]
		case []any:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(current) {
				return nil, d.fail(strings.Join(path, "."), "invalid selection")
			}
			value = current[i]
		default:
			return nil, d.fail(strings.Join(path, "."), "invalid selection")
		}
	}
	if text, ok := value.(string); ok {
		resolved, err := interpolate(text, d.environment)
		if err != nil {
			return nil, d.fail(strings.Join(path, "."), err.Error())
		}
		return resolved, nil
	}
	return value, nil
}
func (d *document) stringAt(path ...string) (string, error) {
	value, err := d.valueAt(path...)
	if err != nil {
		return "", err
	}
	text, ok := value.(string)
	if !ok {
		return "", d.fail(strings.Join(path, "."), "expected string")
	}
	return text, nil
}
