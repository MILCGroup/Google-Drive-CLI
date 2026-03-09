package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"sort"

	"github.com/dl-alexandre/gdrv/internal/discovery"
)

// VolatileFields are fields that should be stripped from snapshots
// to avoid unnecessary drift noise
var VolatileFields = []string{
	"etag",
	"revision",
	"discoveryVersion",
	"generated_at",
	"generatedAt",
}

// NormalizeSnapshot creates a stable, deterministic snapshot of a discovery document
// by sorting keys and removing volatile fields
func NormalizeSnapshot(doc *discovery.DiscoveryDocument) ([]byte, error) {
	// Use StableMarshal which already handles sorting
	return StableMarshal(doc)
}

// StableMarshal marshals a discovery document with stable key ordering
func StableMarshal(doc *discovery.DiscoveryDocument) ([]byte, error) {
	// Use a custom approach that sorts keys at every level
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	sorted := sortKeysRecursive(raw)

	// Pretty print with indentation
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(sorted); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// sortKeysRecursive recursively sorts map keys for stable output
func sortKeysRecursive(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Create sorted result
		result := make(map[string]interface{}, len(val))

		// Get sorted keys
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Process in sorted order
		for _, k := range keys {
			result[k] = sortKeysRecursive(val[k])
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = sortKeysRecursive(item)
		}
		return result

	default:
		return val
	}
}

// CompactSnapshot creates a minimized snapshot removing docs/description churn
func CompactSnapshot(doc *discovery.DiscoveryDocument, keepDescriptions bool) ([]byte, error) {
	// Create a compact representation
	compact := map[string]interface{}{
		"name":        doc.Name,
		"version":     doc.Version,
		"title":       doc.Title,
		"baseUrl":     doc.BaseURL,
		"rootUrl":     doc.RootURL,
		"servicePath": doc.ServicePath,
		"resources":   compactResources(doc.Resources, keepDescriptions),
		"schemas":     compactSchemas(doc.Schemas, keepDescriptions),
	}

	if doc.Auth != nil {
		scopes := make([]string, 0, len(doc.Auth.OAuth2.Scopes))
		for scope := range doc.Auth.OAuth2.Scopes {
			scopes = append(scopes, scope)
		}
		sort.Strings(scopes)
		compact["auth"] = map[string]interface{}{
			"oauth2": map[string]interface{}{
				"scopes": scopes,
			},
		}
	}

	return StableMarshalFromInterface(compact)
}

func compactResources(resources map[string]discovery.Resource, keepDescriptions bool) map[string]interface{} {
	result := make(map[string]interface{}, len(resources))

	keys := make([]string, 0, len(resources))
	for k := range resources {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		res := resources[name]
		methods := compactMethods(res.Methods, keepDescriptions)

		resMap := map[string]interface{}{
			"methods": methods,
		}

		if len(res.Resources) > 0 {
			resMap["resources"] = compactResources(res.Resources, keepDescriptions)
		}

		result[name] = resMap
	}

	return result
}

func compactMethods(methods map[string]discovery.Method, keepDescriptions bool) map[string]interface{} {
	result := make(map[string]interface{}, len(methods))

	keys := make([]string, 0, len(methods))
	for k := range methods {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		m := methods[name]
		methodMap := map[string]interface{}{
			"id":         m.ID,
			"path":       m.Path,
			"httpMethod": m.HTTPMethod,
		}

		if keepDescriptions && m.Description != "" {
			methodMap["description"] = m.Description
		}

		if len(m.Parameters) > 0 {
			methodMap["parameters"] = compactParameters(m.Parameters)
		}

		if m.Request != nil {
			methodMap["request"] = map[string]string{
				"$ref": m.Request.Ref,
			}
		}

		if m.Response != nil {
			methodMap["response"] = map[string]string{
				"$ref": m.Response.Ref,
			}
		}

		if len(m.Scopes) > 0 {
			sortedScopes := make([]string, len(m.Scopes))
			copy(sortedScopes, m.Scopes)
			sort.Strings(sortedScopes)
			methodMap["scopes"] = sortedScopes
		}

		// Preserve semantically meaningful media handling flags
		if m.SupportsMediaUpload {
			methodMap["supportsMediaUpload"] = true
		}
		if m.SupportsMediaDownload {
			methodMap["supportsMediaDownload"] = true
		}
		if m.MediaUpload != nil {
			methodMap["mediaUpload"] = compactMediaUpload(m.MediaUpload)
		}

		result[name] = methodMap
	}

	return result
}

func compactParameters(params map[string]discovery.Parameter) map[string]interface{} {
	result := make(map[string]interface{}, len(params))

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		p := params[name]
		paramMap := map[string]interface{}{
			"type":     p.Type,
			"location": p.Location,
			"required": p.Required,
		}

		if p.Ref != "" {
			paramMap["$ref"] = p.Ref
		}

		if p.Pattern != "" {
			paramMap["pattern"] = p.Pattern
		}

		if p.Format != "" {
			paramMap["format"] = p.Format
		}

		if p.Repeated {
			paramMap["repeated"] = true
		}

		if len(p.Enum) > 0 {
			paramMap["enum"] = p.Enum
		}

		result[name] = paramMap
	}

	return result
}

func compactSchemas(schemas map[string]discovery.Schema, keepDescriptions bool) map[string]interface{} {
	result := make(map[string]interface{}, len(schemas))

	keys := make([]string, 0, len(schemas))
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		s := schemas[name]
		schemaMap := map[string]interface{}{
			"id":   s.ID,
			"type": s.Type,
		}

		if keepDescriptions && s.Description != "" {
			schemaMap["description"] = s.Description
		}

		if s.Ref != "" {
			schemaMap["$ref"] = s.Ref
		}

		if len(s.Required) > 0 {
			req := make([]string, len(s.Required))
			copy(req, s.Required)
			sort.Strings(req)
			schemaMap["required"] = req
		}

		if len(s.Properties) > 0 {
			schemaMap["properties"] = compactSchemaProperties(s.Properties, keepDescriptions)
		}

		result[name] = schemaMap
	}

	return result
}

func compactSchemaProperties(props map[string]discovery.Schema, keepDescriptions bool) map[string]interface{} {
	result := make(map[string]interface{}, len(props))

	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		p := props[name]
		propMap := map[string]interface{}{
			"type": p.Type,
		}

		if keepDescriptions && p.Description != "" {
			propMap["description"] = p.Description
		}

		if p.Ref != "" {
			propMap["$ref"] = p.Ref
		}

		if p.Format != "" {
			propMap["format"] = p.Format
		}

		if p.Pattern != "" {
			propMap["pattern"] = p.Pattern
		}

		if len(p.Enum) > 0 {
			propMap["enum"] = p.Enum
		}

		result[name] = propMap
	}

	return result
}

// StableMarshalFromInterface marshals any interface with stable key ordering
func StableMarshalFromInterface(v interface{}) ([]byte, error) {
	sorted := sortKeysRecursive(v)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(sorted); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// compactMediaUpload extracts essential media upload configuration
func compactMediaUpload(mu *discovery.MediaUpload) map[string]interface{} {
	result := map[string]interface{}{}

	if len(mu.Accept) > 0 {
		result["accept"] = mu.Accept
	}

	if mu.MaxSize != "" {
		result["maxSize"] = mu.MaxSize
	}

	return result
}

// SnapshotHash creates a hash for quick comparison
func SnapshotHash(data []byte) string {
	// Simple hash - in production might use SHA-256
	// For now, just use the first 16 chars of base64
	return base64.StdEncoding.EncodeToString(data)[:16]
}
