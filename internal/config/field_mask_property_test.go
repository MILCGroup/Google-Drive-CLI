package config

import (
	"strings"
	"testing"
)

// Property 7: Field Mask Configuration
// Validates: Requirements 20.1, 20.2, 20.3
// Property: Field masks must be properly formatted and contain expected baseline fields

func TestProperty_FieldMaskConfiguration_BaselineFields(t *testing.T) {
	// Property: Each preset must contain baseline essential fields
	tests := []struct {
		preset             FieldMaskPreset
		requiredFields     []string
		includeExportLinks bool
	}{
		{
			FieldMaskMinimal,
			[]string{"id", "name", "mimeType"},
			false,
		},
		{
			FieldMaskStandard,
			[]string{"id", "name", "mimeType", "size", "createdTime", "modifiedTime", "parents", "trashed", "webViewLink", "webContentLink", "resourceKey", "capabilities"},
			false,
		},
		{
			FieldMaskStandard,
			[]string{"id", "name", "mimeType", "size", "createdTime", "modifiedTime", "parents", "trashed", "webViewLink", "webContentLink", "resourceKey", "capabilities", "exportLinks"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.preset)+"_exportLinks_"+boolToString(tt.includeExportLinks), func(t *testing.T) {
			mask := GetFieldMask(tt.preset, tt.includeExportLinks)

			// Property: Field mask must be properly formatted
			if !strings.HasPrefix(mask, "files(") || !strings.HasSuffix(mask, ")") {
				t.Errorf("Field mask format invalid: %s (must start with 'files(' and end with ')')", mask)
			}

			// Property: All required fields must be present
			fieldsPart := strings.TrimPrefix(strings.TrimSuffix(mask, ")"), "files(")
			fields := strings.Split(fieldsPart, ",")

			for _, requiredField := range tt.requiredFields {
				found := false
				for _, field := range fields {
					field = strings.TrimSpace(field)
					if strings.Contains(field, requiredField) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Required field '%s' not found in mask: %s", requiredField, mask)
				}
			}

			// Property: exportLinks should only be present when requested
			hasExportLinks := strings.Contains(mask, "exportLinks")
			if hasExportLinks != tt.includeExportLinks {
				t.Errorf("exportLinks presence mismatch: expected %v, got %v in mask: %s",
					tt.includeExportLinks, hasExportLinks, mask)
			}
		})
	}
}

func TestProperty_FieldMaskConfiguration_AdditiveFields(t *testing.T) {
	// Property: When includeExportLinks is true, it should add to baseline rather than replace
	// This validates requirement 20.2: "add delta fields to the baseline mask rather than replacing it"

	allPresets := []FieldMaskPreset{FieldMaskMinimal, FieldMaskStandard, FieldMaskFull}

	for _, preset := range allPresets {
		t.Run(string(preset), func(t *testing.T) {
			withoutExportLinks := GetFieldMask(preset, false)
			withExportLinks := GetFieldMask(preset, true)

			// Extract fields
			fieldsWithout := extractFieldsFromMask(withoutExportLinks)
			fieldsWith := extractFieldsFromMask(withExportLinks)

			// Property: With exportLinks should have all fields from without, plus exportLinks
			if len(fieldsWith) != len(fieldsWithout)+1 {
				t.Errorf("Expected exactly one additional field, got %d vs %d",
					len(fieldsWith), len(fieldsWithout))
			}

			// All fields from without should be present in with
			fieldMap := make(map[string]bool)
			for _, field := range fieldsWith {
				fieldMap[field] = true
			}

			for _, field := range fieldsWithout {
				if !fieldMap[field] {
					t.Errorf("Field '%s' from base mask missing in extended mask", field)
				}
			}

			// exportLinks should be present in with but not in without
			if !containsField(fieldsWith, "exportLinks") {
				t.Errorf("exportLinks not found in extended mask: %s", withExportLinks)
			}

			if containsField(fieldsWithout, "exportLinks") {
				t.Errorf("exportLinks unexpectedly found in base mask: %s", withoutExportLinks)
			}
		})
	}
}

// Helper functions

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func extractFieldsFromMask(mask string) []string {
	if !strings.HasPrefix(mask, "files(") || !strings.HasSuffix(mask, ")") {
		return nil
	}
	fieldsPart := strings.TrimPrefix(strings.TrimSuffix(mask, ")"), "files(")
	return strings.Split(fieldsPart, ",")
}

func containsField(fields []string, targetField string) bool {
	for _, field := range fields {
		if strings.TrimSpace(field) == targetField {
			return true
		}
	}
	return false
}
