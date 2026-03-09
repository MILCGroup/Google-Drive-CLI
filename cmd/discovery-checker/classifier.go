package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/discovery"
)

// ChangeSeverity represents the risk level of a change
type ChangeSeverity string

const (
	// SeverityAdditive: new optional fields, new methods - safe to auto-merge
	SeverityAdditive ChangeSeverity = "additive"
	// SeverityRisky: type changes, new required params - needs owner review
	SeverityRisky ChangeSeverity = "risky"
	// SeverityBreaking: removals, incompatible changes - blocks pipeline
	SeverityBreaking ChangeSeverity = "breaking"
)

// ChangeCategory provides context about what changed
type ChangeCategory string

const (
	CategoryType      ChangeCategory = "type"
	CategoryField     ChangeCategory = "field"
	CategoryMethod    ChangeCategory = "method"
	CategoryResource  ChangeCategory = "resource"
	CategoryParameter ChangeCategory = "parameter"
	CategoryPath      ChangeCategory = "path"
	CategoryHTTP      ChangeCategory = "http"
	CategoryRequest   ChangeCategory = "request"
	CategoryResponse  ChangeCategory = "response"
	CategoryAuth      ChangeCategory = "auth"
	CategoryScope     ChangeCategory = "scope"
	CategoryEnum      ChangeCategory = "enum"
	CategoryFormat    ChangeCategory = "format"
)

// Change represents a detected drift with full context
type Change struct {
	Severity    ChangeSeverity `json:"severity"`
	Category    ChangeCategory `json:"category"`
	Description string         `json:"description"`
	API         string         `json:"api"`
	Resource    string         `json:"resource,omitempty"`
	Method      string         `json:"method,omitempty"`
	Field       string         `json:"field,omitempty"`
	// Additional metadata for detailed reporting
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

// ChangeReport aggregates all changes with summary statistics
type ChangeReport struct {
	API               string   `json:"api"`
	Version           string   `json:"version"`
	TotalChanges      int      `json:"total_changes"`
	AdditiveChanges   int      `json:"additive_changes"`
	RiskyChanges      int      `json:"risky_changes"`
	BreakingChanges   int      `json:"breaking_changes"`
	Changes           []Change `json:"changes"`
	HasBreakingChange bool     `json:"has_breaking_change"`
}

// RefinedClassifier implements sophisticated change detection with edge case handling
type RefinedClassifier struct {
	// Enum expansion policy: true = additive, false = risky
	EnumExpansionIsAdditive bool
	// Format change threshold: int32->int64 = risky, int32->string = breaking
	StrictFormatChecking bool
}

// NewRefinedClassifier creates a classifier with conservative defaults
// Enum expansion defaults to risky (not additive) because consumers may validate enum values
func NewRefinedClassifier() *RefinedClassifier {
	return &RefinedClassifier{
		EnumExpansionIsAdditive: false, // Conservative default: enum expansion is risky
		StrictFormatChecking:    false, // Format changes are risky, not breaking by default
	}
}

// ClassifyChanges compares two discovery documents and produces a detailed report
func (c *RefinedClassifier) ClassifyChanges(apiName string, old, new *discovery.DiscoveryDocument) *ChangeReport {
	report := &ChangeReport{
		API:     apiName,
		Version: new.Version,
		Changes: []Change{},
	}

	// Compare top-level metadata
	report.Changes = append(report.Changes, c.compareMetadata(old, new)...)

	// Compare authentication
	report.Changes = append(report.Changes, c.compareAuth(old, new)...)

	// Compare schemas (types)
	for name, newSchema := range new.Schemas {
		if oldSchema, exists := old.Schemas[name]; exists {
			report.Changes = append(report.Changes, c.compareSchema(name, oldSchema, newSchema)...)
		} else {
			report.Changes = append(report.Changes, Change{
				Severity:    SeverityAdditive,
				Category:    CategoryType,
				Description: fmt.Sprintf("New type added: %s", name),
				API:         apiName,
			})
		}
	}

	// Check for removed schemas
	for name := range old.Schemas {
		if _, exists := new.Schemas[name]; !exists {
			report.Changes = append(report.Changes, Change{
				Severity:    SeverityBreaking,
				Category:    CategoryType,
				Description: fmt.Sprintf("Type removed: %s", name),
				API:         apiName,
			})
		}
	}

	// Compare resources and methods
	for resName, newResource := range new.Resources {
		oldResource, exists := old.Resources[resName]
		if !exists {
			report.Changes = append(report.Changes, Change{
				Severity:    SeverityAdditive,
				Category:    CategoryResource,
				Description: fmt.Sprintf("New resource added: %s", resName),
				API:         apiName,
			})
			continue
		}

		report.Changes = append(report.Changes, c.compareResource(apiName, resName, oldResource, newResource)...)
	}

	// Check for removed resources
	for resName := range old.Resources {
		if _, exists := new.Resources[resName]; !exists {
			report.Changes = append(report.Changes, Change{
				Severity:    SeverityBreaking,
				Category:    CategoryResource,
				Description: fmt.Sprintf("Resource removed: %s", resName),
				API:         apiName,
			})
		}
	}

	// Calculate summary statistics
	c.summarize(report)

	return report
}

func (c *RefinedClassifier) compareMetadata(old, new *discovery.DiscoveryDocument) []Change {
	var changes []Change

	// Check for base URL changes (risky - could affect all requests)
	if old.BaseURL != new.BaseURL {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryPath,
			Description: fmt.Sprintf("Base URL changed: %s -> %s", old.BaseURL, new.BaseURL),
			OldValue:    old.BaseURL,
			NewValue:    new.BaseURL,
		})
	}

	// Service path changes
	if old.ServicePath != new.ServicePath {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryPath,
			Description: fmt.Sprintf("Service path changed: %s -> %s", old.ServicePath, new.ServicePath),
			OldValue:    old.ServicePath,
			NewValue:    new.ServicePath,
		})
	}

	return changes
}

func (c *RefinedClassifier) compareAuth(old, new *discovery.DiscoveryDocument) []Change {
	var changes []Change

	// Check for OAuth scope changes
	if old.Auth != nil && new.Auth != nil {
		oldScopes := make(map[string]bool)
		for scope := range old.Auth.OAuth2.Scopes {
			oldScopes[scope] = true
		}

		newScopes := make(map[string]bool)
		for scope := range new.Auth.OAuth2.Scopes {
			newScopes[scope] = true
		}

		// Check for new scopes
		for scope := range newScopes {
			if !oldScopes[scope] {
				changes = append(changes, Change{
					Severity:    SeverityAdditive,
					Category:    CategoryScope,
					Description: fmt.Sprintf("New OAuth scope available: %s", scope),
					NewValue:    scope,
				})
			}
		}

		// Check for removed scopes - this is risky, not breaking
		// (existing tokens still work, but new auth might need different scopes)
		for scope := range oldScopes {
			if !newScopes[scope] {
				changes = append(changes, Change{
					Severity:    SeverityRisky,
					Category:    CategoryScope,
					Description: fmt.Sprintf("OAuth scope removed: %s", scope),
					OldValue:    scope,
				})
			}
		}
	}

	// Check for auth requirement changes
	if old.Auth == nil && new.Auth != nil {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryAuth,
			Description: "Authentication added to previously unauthenticated API",
		})
	}

	if old.Auth != nil && new.Auth == nil {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryAuth,
			Description: "Authentication removed - API may be deprecated",
		})
	}

	return changes
}

func (c *RefinedClassifier) compareSchema(name string, old, new discovery.Schema) []Change {
	var changes []Change

	// Check for type changes
	if old.Type != new.Type {
		severity := c.classifyTypeChange(old.Type, new.Type)
		changes = append(changes, Change{
			Severity:    severity,
			Category:    CategoryType,
			Description: fmt.Sprintf("Type %s changed: %s -> %s", name, old.Type, new.Type),
			Field:       name,
			OldValue:    old.Type,
			NewValue:    new.Type,
		})
		return changes // Type changes are fundamental, don't check deeper
	}

	// Check for $ref changes
	if old.Ref != new.Ref {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryType,
			Description: fmt.Sprintf("Type reference changed in %s: %s -> %s", name, old.Ref, new.Ref),
			Field:       name,
			OldValue:    old.Ref,
			NewValue:    new.Ref,
		})
	}

	// Check required field changes
	oldRequired := make(map[string]bool)
	for _, f := range old.Required {
		oldRequired[f] = true
	}

	newRequired := make(map[string]bool)
	for _, f := range new.Required {
		newRequired[f] = true
	}

	// Fields becoming required
	for field := range newRequired {
		if !oldRequired[field] {
			changes = append(changes, Change{
				Severity:    SeverityRisky, // Risky, not breaking (graceful degradation possible)
				Category:    CategoryField,
				Description: fmt.Sprintf("Field %s became required in type %s", field, name),
				Field:       field,
			})
		}
	}

	// Fields becoming optional (usually safe, but note it)
	for field := range oldRequired {
		if !newRequired[field] {
			changes = append(changes, Change{
				Severity:    SeverityAdditive,
				Category:    CategoryField,
				Description: fmt.Sprintf("Field %s became optional in type %s", field, name),
				Field:       field,
			})
		}
	}

	// Compare properties
	for propName, newProp := range new.Properties {
		if oldProp, exists := old.Properties[propName]; exists {
			changes = append(changes, c.compareProperty(name, propName, oldProp, newProp)...)
		} else {
			// New property - check if it's required
			severity := SeverityAdditive
			if newRequired[propName] {
				severity = SeverityRisky // New required field is risky
			}
			changes = append(changes, Change{
				Severity:    severity,
				Category:    CategoryField,
				Description: fmt.Sprintf("New field added to %s: %s", name, propName),
				Field:       propName,
			})
		}
	}

	// Check for removed properties
	for propName := range old.Properties {
		if _, exists := new.Properties[propName]; !exists {
			changes = append(changes, Change{
				Severity:    SeverityBreaking,
				Category:    CategoryField,
				Description: fmt.Sprintf("Field removed from %s: %s", name, propName),
				Field:       propName,
			})
		}
	}

	// Check additionalProperties changes (schema tightening/relaxing)
	if !reflect.DeepEqual(old.AdditionalProperties, new.AdditionalProperties) {
		// Schema became more restrictive
		if old.AdditionalProperties != nil && new.AdditionalProperties == nil {
			changes = append(changes, Change{
				Severity:    SeverityRisky,
				Category:    CategoryType,
				Description: fmt.Sprintf("Type %s tightened additionalProperties restrictions", name),
			})
		}
		// Schema became more permissive
		if old.AdditionalProperties == nil && new.AdditionalProperties != nil {
			changes = append(changes, Change{
				Severity:    SeverityAdditive,
				Category:    CategoryType,
				Description: fmt.Sprintf("Type %s relaxed additionalProperties restrictions", name),
			})
		}
	}

	return changes
}

func (c *RefinedClassifier) compareProperty(typeName, propName string, old, new discovery.Schema) []Change {
	var changes []Change

	// Check for type changes
	if old.Type != new.Type {
		severity := c.classifyTypeChange(old.Type, new.Type)
		changes = append(changes, Change{
			Severity:    severity,
			Category:    CategoryField,
			Description: fmt.Sprintf("Type change in %s.%s: %s -> %s", typeName, propName, old.Type, new.Type),
			Field:       propName,
			OldValue:    old.Type,
			NewValue:    new.Type,
		})
		return changes
	}

	// Check for $ref changes
	if old.Ref != new.Ref {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryField,
			Description: fmt.Sprintf("Type reference change in %s.%s: %s -> %s", typeName, propName, old.Ref, new.Ref),
			Field:       propName,
			OldValue:    old.Ref,
			NewValue:    new.Ref,
		})
	}

	// Check for format changes (int32 -> int64, etc.)
	if old.Format != new.Format {
		severity := SeverityRisky
		if c.StrictFormatChecking {
			severity = c.classifyFormatChange(old.Format, new.Format)
		}
		changes = append(changes, Change{
			Severity:    severity,
			Category:    CategoryFormat,
			Description: fmt.Sprintf("Format change in %s.%s: %s -> %s", typeName, propName, old.Format, new.Format),
			Field:       propName,
			OldValue:    old.Format,
			NewValue:    new.Format,
		})
	}

	// Check enum changes with configurable policy
	if len(old.Enum) > 0 || len(new.Enum) > 0 {
		changes = append(changes, c.compareEnums(typeName, propName, old.Enum, new.Enum)...)
	}

	// Check pattern changes
	if old.Pattern != new.Pattern {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryField,
			Description: fmt.Sprintf("Pattern change in %s.%s", typeName, propName),
			Field:       propName,
			OldValue:    old.Pattern,
			NewValue:    new.Pattern,
		})
	}

	return changes
}

func (c *RefinedClassifier) compareEnums(typeName, propName string, old, new []string) []Change {
	var changes []Change

	oldEnums := make(map[string]bool)
	for _, e := range old {
		oldEnums[e] = true
	}

	newEnums := make(map[string]bool)
	for _, e := range new {
		newEnums[e] = true
	}

	// Check for new enum values
	for e := range newEnums {
		if !oldEnums[e] {
			severity := SeverityAdditive
			if !c.EnumExpansionIsAdditive {
				severity = SeverityRisky // If client validates enums strictly
			}
			changes = append(changes, Change{
				Severity:    severity,
				Category:    CategoryEnum,
				Description: fmt.Sprintf("New enum value in %s.%s: %s", typeName, propName, e),
				Field:       propName,
				NewValue:    e,
			})
		}
	}

	// Check for removed enum values - always breaking
	for e := range oldEnums {
		if !newEnums[e] {
			changes = append(changes, Change{
				Severity:    SeverityBreaking,
				Category:    CategoryEnum,
				Description: fmt.Sprintf("Enum value removed from %s.%s: %s", typeName, propName, e),
				Field:       propName,
				OldValue:    e,
			})
		}
	}

	return changes
}

func (c *RefinedClassifier) compareResource(apiName, resName string, old, new discovery.Resource) []Change {
	var changes []Change

	// Compare methods
	for methodName, newMethod := range new.Methods {
		oldMethod, exists := old.Methods[methodName]
		if !exists {
			changes = append(changes, Change{
				Severity:    SeverityAdditive,
				Category:    CategoryMethod,
				Description: fmt.Sprintf("New method added: %s.%s", resName, methodName),
				API:         apiName,
				Resource:    resName,
				Method:      methodName,
			})
			continue
		}

		changes = append(changes, c.compareMethod(apiName, resName, methodName, oldMethod, newMethod)...)
	}

	// Check for removed methods
	for methodName := range old.Methods {
		if _, exists := new.Methods[methodName]; !exists {
			changes = append(changes, Change{
				Severity:    SeverityBreaking,
				Category:    CategoryMethod,
				Description: fmt.Sprintf("Method removed: %s.%s", resName, methodName),
				API:         apiName,
				Resource:    resName,
				Method:      methodName,
			})
		}
	}

	// Recursively compare nested resources
	for nestedName, newNested := range new.Resources {
		oldNested, exists := old.Resources[nestedName]
		if !exists {
			changes = append(changes, Change{
				Severity:    SeverityAdditive,
				Category:    CategoryResource,
				Description: fmt.Sprintf("New nested resource added: %s.%s", resName, nestedName),
				API:         apiName,
				Resource:    resName + "/" + nestedName,
			})
			continue
		}

		changes = append(changes, c.compareResource(apiName, resName+"/"+nestedName, oldNested, newNested)...)
	}

	return changes
}

func (c *RefinedClassifier) compareMethod(apiName, resName, methodName string, old, new discovery.Method) []Change {
	var changes []Change

	// HTTP method changes - breaking
	if old.HTTPMethod != new.HTTPMethod {
		changes = append(changes, Change{
			Severity:    SeverityBreaking,
			Category:    CategoryHTTP,
			Description: fmt.Sprintf("HTTP method changed for %s.%s: %s -> %s", resName, methodName, old.HTTPMethod, new.HTTPMethod),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			OldValue:    old.HTTPMethod,
			NewValue:    new.HTTPMethod,
		})
	}

	// Path template changes - risky (could break URL construction)
	if old.Path != new.Path {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryPath,
			Description: fmt.Sprintf("Path template changed for %s.%s: %s -> %s", resName, methodName, old.Path, new.Path),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			OldValue:    old.Path,
			NewValue:    new.Path,
		})
	}

	// Parameter changes
	for paramName, newParam := range new.Parameters {
		oldParam, exists := old.Parameters[paramName]
		if !exists {
			severity := SeverityAdditive
			if newParam.Required {
				severity = SeverityRisky // New required param is risky
			}
			changes = append(changes, Change{
				Severity:    severity,
				Category:    CategoryParameter,
				Description: fmt.Sprintf("New parameter added to %s.%s: %s", resName, methodName, paramName),
				API:         apiName,
				Resource:    resName,
				Method:      methodName,
				Field:       paramName,
			})
			continue
		}

		changes = append(changes, c.compareParameter(apiName, resName, methodName, paramName, oldParam, newParam)...)
	}

	// Check for removed parameters - breaking
	for paramName := range old.Parameters {
		if _, exists := new.Parameters[paramName]; !exists {
			changes = append(changes, Change{
				Severity:    SeverityBreaking,
				Category:    CategoryParameter,
				Description: fmt.Sprintf("Parameter removed from %s.%s: %s", resName, methodName, paramName),
				API:         apiName,
				Resource:    resName,
				Method:      methodName,
				Field:       paramName,
			})
		}
	}

	// Request type changes
	if old.Request != nil && new.Request != nil {
		if old.Request.Ref != new.Request.Ref || old.Request.Type != new.Request.Type {
			changes = append(changes, Change{
				Severity:    SeverityRisky,
				Category:    CategoryRequest,
				Description: fmt.Sprintf("Request type changed for %s.%s", resName, methodName),
				API:         apiName,
				Resource:    resName,
				Method:      methodName,
			})
		}
	}

	// Response type changes
	if old.Response != nil && new.Response != nil {
		if old.Response.Ref != new.Response.Ref || old.Response.Type != new.Response.Type {
			changes = append(changes, Change{
				Severity:    SeverityRisky,
				Category:    CategoryResponse,
				Description: fmt.Sprintf("Response type changed for %s.%s", resName, methodName),
				API:         apiName,
				Resource:    resName,
				Method:      methodName,
			})
		}
	}

	// Scope changes (per-method)
	if !stringSliceEqual(old.Scopes, new.Scopes) {
		severity := SeverityRisky // Scope changes are always risky minimum

		// Check if scopes were removed
		oldScopeSet := make(map[string]bool)
		for _, s := range old.Scopes {
			oldScopeSet[s] = true
		}
		for _, s := range new.Scopes {
			if !oldScopeSet[s] {
				severity = SeverityRisky // New scope required
				break
			}
		}

		changes = append(changes, Change{
			Severity:    severity,
			Category:    CategoryScope,
			Description: fmt.Sprintf("OAuth scopes changed for %s.%s", resName, methodName),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			OldValue:    old.Scopes,
			NewValue:    new.Scopes,
		})
	}

	return changes
}

func (c *RefinedClassifier) compareParameter(apiName, resName, methodName, paramName string, old, new discovery.Parameter) []Change {
	var changes []Change

	// Required status changes
	if !old.Required && new.Required {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryParameter,
			Description: fmt.Sprintf("Parameter %s became required in %s.%s", paramName, resName, methodName),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			Field:       paramName,
		})
	}

	// Type changes
	if old.Type != new.Type {
		severity := c.classifyTypeChange(old.Type, new.Type)
		changes = append(changes, Change{
			Severity:    severity,
			Category:    CategoryParameter,
			Description: fmt.Sprintf("Parameter %s type changed in %s.%s: %s -> %s", paramName, resName, methodName, old.Type, new.Type),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			Field:       paramName,
			OldValue:    old.Type,
			NewValue:    new.Type,
		})
	}

	// Format changes
	if old.Format != new.Format {
		severity := SeverityRisky
		if c.StrictFormatChecking {
			severity = c.classifyFormatChange(old.Format, new.Format)
		}
		changes = append(changes, Change{
			Severity:    severity,
			Category:    CategoryFormat,
			Description: fmt.Sprintf("Parameter %s format changed in %s.%s: %s -> %s", paramName, resName, methodName, old.Format, new.Format),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			Field:       paramName,
			OldValue:    old.Format,
			NewValue:    new.Format,
		})
	}

	// Location changes (path -> query or vice versa) - risky
	if old.Location != new.Location {
		changes = append(changes, Change{
			Severity:    SeverityRisky,
			Category:    CategoryParameter,
			Description: fmt.Sprintf("Parameter %s location changed in %s.%s: %s -> %s", paramName, resName, methodName, old.Location, new.Location),
			API:         apiName,
			Resource:    resName,
			Method:      methodName,
			Field:       paramName,
			OldValue:    old.Location,
			NewValue:    new.Location,
		})
	}

	// Enum changes for parameters
	if len(old.Enum) > 0 || len(new.Enum) > 0 {
		changes = append(changes, c.compareEnums(resName+"/"+methodName, paramName, old.Enum, new.Enum)...)
	}

	return changes
}

// classifyTypeChange determines severity based on type change rules
func (c *RefinedClassifier) classifyTypeChange(old, new string) ChangeSeverity {
	// Same type - no change
	if old == new {
		return SeverityAdditive
	}

	// Integer widening is risky but usually safe
	if isIntegerType(old) && isIntegerType(new) {
		return SeverityRisky
	}

	// Number <-> Integer is risky
	if (old == "number" && new == "integer") || (old == "integer" && new == "number") {
		return SeverityRisky
	}

	// String <-> Integer/Number is breaking (parsing changes)
	if (old == "string" && (isNumericType(new))) ||
		(new == "string" && isNumericType(old)) {
		return SeverityBreaking
	}

	// Array <-> Object is breaking
	if (old == "array" && new == "object") || (old == "object" && new == "array") {
		return SeverityBreaking
	}

	// Default to risky when in doubt
	return SeverityRisky
}

// classifyFormatChange determines severity based on format change rules
func (c *RefinedClassifier) classifyFormatChange(old, new string) ChangeSeverity {
	// Empty format changes are usually safe
	if old == "" || new == "" {
		return SeverityAdditive
	}

	// int32 -> int64 is widening, usually safe
	if old == "int32" && new == "int64" {
		return SeverityRisky
	}

	// int64 -> int32 is narrowing, risky
	if old == "int64" && new == "int32" {
		return SeverityRisky
	}

	// Date format changes are risky (parsing changes)
	if strings.Contains(old, "date") || strings.Contains(new, "date") {
		return SeverityRisky
	}

	// Default to risky for format changes
	return SeverityRisky
}

func isIntegerType(t string) bool {
	return t == "integer"
}

func isNumericType(t string) bool {
	return t == "integer" || t == "number"
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (c *RefinedClassifier) summarize(report *ChangeReport) {
	report.TotalChanges = len(report.Changes)

	for _, change := range report.Changes {
		switch change.Severity {
		case SeverityAdditive:
			report.AdditiveChanges++
		case SeverityRisky:
			report.RiskyChanges++
		case SeverityBreaking:
			report.BreakingChanges++
			report.HasBreakingChange = true
		}
	}
}
