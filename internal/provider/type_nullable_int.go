package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NullableInt handles special JSON marshaling for nullable integer fields in the Better Uptime API.
// It's used for fields like ssl_expiration and domain_expiration that can be:
// - Unset: field is omitted from JSON (when Value is nil and ExplicitNull is false)
// - Set to null: field is sent as null in JSON (when ExplicitNull is true)
// - Set to a value: field is sent as that value (when Value is set)
//
// Example usage in resource_monitor.go:
//   - For ssl_expiration and domain_expiration fields that can be set to -1 to disable checks
//   - When set to -1, the field is sent as null in the API
//   - When unset, the field is omitted from the API request
//   - When set to a value (1, 2, 3, 7, 14, 30, 60), that value is sent to the API
type NullableInt struct {
	// Value is the actual integer value. When nil and ExplicitNull is false, the field is omitted.
	Value *int
	// ExplicitNull indicates whether the field should be marshaled as null (set to -1 in Terraform).
	ExplicitNull bool
}

// MarshalJSON implements json.Marshaler interface.
// Used when sending data to the Better Uptime API:
// - If ExplicitNull is true (Terraform value was -1), returns "null"
// - If Value is set (Terraform value was 1-60), returns that value
// - If neither (field was unset), returns nil (field is omitted)
func (n NullableInt) MarshalJSON() ([]byte, error) {
	if n.ExplicitNull {
		return []byte("null"), nil
	}
	if n.Value != nil {
		return json.Marshal(*n.Value)
	}
	return nil, nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
// Used when receiving data from the Better Uptime API:
// - When the JSON value is "null", sets ExplicitNull to true (will be shown as -1 in Terraform)
// - Otherwise, unmarshals the value into Value (will be shown as that value in Terraform)
func (n *NullableInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Value = nil
		n.ExplicitNull = true
		return nil
	}
	var v int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	n.Value = &v
	n.ExplicitNull = false
	return nil
}

// NullableIntFromResourceData creates a NullableInt from a Terraform resource field.
// Used in monitorCreate and monitorUpdate to handle ssl_expiration and domain_expiration fields.
// Parameters:
//   - d: The Terraform resource data
//   - key: The field name (e.g., "ssl_expiration" or "domain_expiration")
//   - terraformNullValue: The value that represents null in Terraform (e.g., -1)
//
// Example from resource_monitor.go:
//
//	in.SSLExpiration = NullableIntFromResourceData(d, "ssl_expiration", -1)
func NullableIntFromResourceData(d *schema.ResourceData, key string, terraformNullValue int) *NullableInt {
	if v, ok := d.GetOk(key); ok {
		ni := &NullableInt{}
		if v == nil {
			return nil
		}
		val := v.(int)
		if val == terraformNullValue {
			ni.ExplicitNull = true
		} else {
			ni.Value = &val
		}
		return ni
	}
	return nil
}

// SetNullableIntResourceData sets a Terraform resource field from a NullableInt.
// Used in monitorCopyAttrs to handle ssl_expiration and domain_expiration fields.
// Parameters:
//   - d: The Terraform resource data
//   - key: The field name (e.g., "ssl_expiration" or "domain_expiration")
//   - terraformNullValue: The value to use for null in Terraform (e.g., -1)
//   - nullableInt: The NullableInt value from the API
//
// Example from resource_monitor.go:
//
//	SetNullableIntResourceData(d, "ssl_expiration", -1, in.SSLExpiration)
func SetNullableIntResourceData(d *schema.ResourceData, key string, terraformNullValue int, nullableInt *NullableInt) {
	if nullableInt == nil || nullableInt.ExplicitNull {
		d.Set(key, terraformNullValue)
	} else if nullableInt.Value != nil {
		d.Set(key, *nullableInt.Value)
	}
}
