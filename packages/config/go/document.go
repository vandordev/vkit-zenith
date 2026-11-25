package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

type shape map[string]shape

var serviceShape = shape{"host": nil, "port": nil, "readTimeoutSeconds": nil, "writeTimeoutSeconds": nil, "idleTimeoutSeconds": nil, "shutdownTimeoutSeconds": nil, "docs": {"basicAuth": {"enabled": nil, "username": nil, "password": nil}}}
var databaseShape = shape{"dsn": nil, "maxOpenConnections": nil, "connectionMaxLifetimeSeconds": nil, "connectionMaxIdleTimeSeconds": nil, "connectTimeoutSeconds": nil}
var workerShape = shape{"shutdownTimeoutSeconds": nil, "maxWorkers": nil}
var storageShape = shape{"endpoint": nil, "region": nil, "bucket": nil, "accessKeyId": nil, "secretAccessKey": nil, "presignTtlSeconds": nil}

func serviceWith(extra shape) shape {
	out := make(shape, len(serviceShape)+len(extra))
	for k, v := range serviceShape {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

var grammar = shape{
	"db":         {"dsn": nil, "maxOpenConnections": nil, "connectionMaxLifetimeSeconds": nil, "connectionMaxIdleTimeSeconds": nil, "connectTimeoutSeconds": nil},
	"worker":     workerShape,
	"storage":    storageShape,
	"encryption": {"keys": {"[]": nil}},
	"app":        {"environment": nil, "logLevel": nil},
	"web":        {"host": nil, "port": nil, "apiBaseUrl": nil},
	"bo":         {"host": nil, "port": nil, "boApiBaseUrl": nil},
	"boApi":      {"host": nil, "port": nil, "readTimeoutSeconds": nil, "writeTimeoutSeconds": nil, "idleTimeoutSeconds": nil, "shutdownTimeoutSeconds": nil, "docs": {"basicAuth": {"enabled": nil, "username": nil, "password": nil}}, "boOrigin": nil, "apiOrigin": nil, "development": nil, "trustedProxyCIDRs": {"[]?": nil}, "csrfSecret": nil},
	"api":        serviceWith(shape{}),
	"webApi":     serviceWith(shape{"identity": {"authOrigin": nil, "development": nil, "trustedProxyCIDRs": {"[]?": nil}, "csrfKey": nil, "limiterKey": nil, "email": {"from": nil, "resendKey": nil}}}),
}
var digits = regexp.MustCompile(`^[0-9]+$`)
var decimal = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)
var yamlBool = regexp.MustCompile(`^(?:true|True|TRUE|false|False|FALSE)$`)
var yamlNull = regexp.MustCompile(`^(?:null|Null|NULL|~)?$`)
var yamlNonfinite = regexp.MustCompile(`^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$`)
var yamlOctal = regexp.MustCompile(`^0o[0-7]+$`)
var yamlHex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)

type document struct {
	root          map[string]any
	environment   map[string]string
	file, runtime string
}

func (d *document) fail(field, reason string) error {
	return fmt.Errorf("Configuration %s [%s] %s: %s", d.file, d.runtime, field, reason)
}

func loadDocument(options Options, runtime string) (*document, error) {
	root, err := ApplicationRoot(options.Root)
	if err != nil {
		return nil, err
	}
	d := &document{file: filepath.Join(root, "config", "config.yaml"), runtime: runtime, environment: options.Environment}
	if d.environment == nil {
		d.environment = Environment()
	}
	source, err := os.ReadFile(d.file)
	if err != nil {
		return nil, d.fail("document", "cannot read document")
	}
	// yaml.v3 accepts 1.1 directives, not 1.2. Reject 1.1 and strip only the
	// supported 1.2 directive; scalar normalization below owns 1.2 semantics.
	lines := strings.Split(string(source), "\n")
	directives := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "%YAML") {
			directives++
			if directives > 1 {
				return nil, d.fail("document", "duplicate YAML directive")
			}
			if strings.TrimSpace(strings.SplitN(line, "#", 2)[0]) != "%YAML 1.2" {
				return nil, d.fail("document", "expected YAML 1.2")
			}
			lines[i] = "" // preserve line positions
		}
	}
	decoder := yaml.NewDecoder(bytes.NewBufferString(strings.Join(lines, "\n")))
	var node, extra yaml.Node
	if decoder.Decode(&node) != nil || len(node.Content) != 1 {
		return nil, d.fail("document", "invalid YAML document")
	}
	if decoder.Decode(&extra) != io.EOF {
		return nil, d.fail("document", "multiple or invalid documents")
	}
	value, err := d.normalize(node.Content[0], grammar, "document")
	if err != nil {
		return nil, err
	}
	d.root = value.(map[string]any)
	return d, nil
}
func (d *document) normalize(node *yaml.Node, expected shape, path string) (any, error) {
	if node.Anchor != "" || node.Kind == yaml.AliasNode || node.Style&yaml.TaggedStyle != 0 {
		return nil, d.fail(path, "aliases, anchors and explicit tags are forbidden")
	}
	if expected != nil {
		if child, sequence := expected["[]8"]; sequence {
			if node.Kind != yaml.SequenceNode || len(node.Content) == 0 || len(node.Content) > 8 {
				return nil, d.fail(path, "expected one to eight keys")
			}
			out := make([]any, 0, len(node.Content))
			for i, item := range node.Content {
				value, err := d.normalize(item, child, fmt.Sprintf("%s.%d", path, i))
				if err != nil {
					return nil, err
				}
				out = append(out, value)
			}
			return out, nil
		}
		if child, sequence := expected["[]?"]; sequence {
			if node.Kind != yaml.SequenceNode || len(node.Content) > 64 {
				return nil, d.fail(path, "expected bounded sequence")
			}
			out := make([]any, 0, len(node.Content))
			for i, item := range node.Content {
				value, err := d.normalize(item, child, fmt.Sprintf("%s.%d", path, i))
				if err != nil {
					return nil, err
				}
				out = append(out, value)
			}
			return out, nil
		}
		if child, sequence := expected["[]"]; sequence {
			if node.Kind != yaml.SequenceNode || len(node.Content) == 0 {
				return nil, d.fail(path, "expected nonempty sequence")
			}
			out := make([]any, 0, len(node.Content))
			for i, item := range node.Content {
				value, err := d.normalize(item, child, fmt.Sprintf("%s.%d", path, i))
				if err != nil {
					return nil, err
				}
				out = append(out, value)
			}
			return out, nil
		}
		if node.Kind != yaml.MappingNode {
			return nil, d.fail(path, "expected mapping")
		}
		out := make(map[string]any)
		for i := 0; i < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Anchor != "" || key.Style&yaml.TaggedStyle != 0 || strings.Contains(key.Value, "${") || key.Value == "<<" {
				return nil, d.fail(path, "invalid or templated key")
			}
			child, ok := expected[key.Value]
			if !ok {
				return nil, d.fail(path, "unknown key")
			}
			if _, exists := out[key.Value]; exists {
				return nil, d.fail(path, "duplicate key")
			}
			normalized, err := d.normalize(value, child, path+"."+key.Value)
			if err != nil {
				return nil, err
			}
			out[key.Value] = normalized
		}
		if len(out) != len(expected) {
			return nil, d.fail(path, "missing key")
		}
		return out, nil
	}
	if node.Kind != yaml.ScalarNode {
		return nil, d.fail(path, "expected scalar")
	}
	if _, err := interpolate(node.Value, nil); err != nil {
		return nil, d.fail(path, "malformed template")
	}
	if node.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle|yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
		return node.Value, nil
	}
	// Resolve the same core scalar grammar as the JS yaml 1.2 parser, not
	// yaml.v3's implicit 1.1 tags (timestamps, binary and underscore numbers).
	text := node.Value
	if yamlNull.MatchString(text) {
		return nil, nil
	}
	if yamlBool.MatchString(text) {
		return strings.EqualFold(text, "true"), nil
	}
	if yamlNonfinite.MatchString(text) {
		return nil, d.fail(path, "nonfinite number")
	}
	if decimal.MatchString(text) {
		n, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, d.fail(path, "invalid number")
		}
		return n, nil
	}
	if yamlOctal.MatchString(text) || yamlHex.MatchString(text) {
		base := 16
		if yamlOctal.MatchString(text) {
			base = 8
		}
		n, err := strconv.ParseUint(text[2:], base, 64)
		if err != nil {
			return nil, d.fail(path, "invalid number")
		}
		return float64(n), nil
	}
	return text, nil
}
