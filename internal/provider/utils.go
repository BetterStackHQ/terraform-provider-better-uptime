package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func suppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	// If either value is empty, don't suppress the diff
	if old == "" || new == "" {
		return false
	}

	// Try to normalize both JSON strings
	normalizedOld, err := normalizeJSON(old)
	if err != nil {
		return false
	}

	normalizedNew, err := normalizeJSON(new)
	if err != nil {
		return false
	}

	return normalizedOld == normalizedNew
}

func normalizeJSON(jsonString string) (string, error) {
	// Trim whitespace
	jsonString = strings.TrimSpace(jsonString)

	// If the string is already an object/array, parse it directly
	var result interface{}
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		// If it's not valid JSON, try wrapping it in quotes to make it a valid JSON string
		jsonString = `"` + strings.Replace(jsonString, `"`, `\"`, -1) + `"`
		err = json.Unmarshal([]byte(jsonString), &result)
		if err != nil {
			return "", err
		}
	}

	// Marshal back to string in a normalized format
	normalizedBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(normalizedBytes), nil
}
