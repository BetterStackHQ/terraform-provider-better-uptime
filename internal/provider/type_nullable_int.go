package provider

import (
	"encoding/json"
)

// NullableInt is a custom type that handles special JSON marshaling for nullable integer fields.
// It is designed to handle three states:
// 1. Unset (field is omitted from JSON)
// 2. ExplicitNull = true (field is sent as null in JSON, i.e. set to -1 in Terraform)
// 3. Value is set (field is sent as that value in JSON)
//
// Example usage:
//
//	type Monitor struct {
//	    SSLExpiration *NullableInt `json:"ssl_expiration,omitempty"`
//	}
//
//	// When unset in Terraform:
//	// JSON: {} (field is omitted)
//	// Terraform: field is not set
//
//	// When set to -1 in Terraform:
//	// JSON: {"ssl_expiration": null}
//	// Terraform: ssl_expiration = -1
//
//	// When set to 7 in Terraform:
//	// JSON: {"ssl_expiration": 7}
//	// Terraform: ssl_expiration = 7
type NullableInt struct {
	// Value is the actual integer value. When nil and ExplicitNull is false, the field is omitted.
	Value *int
	// ExplicitNull indicates whether the field should be marshaled as null (set to -1 in Terraform).
	ExplicitNull bool
}

// MarshalJSON implements json.Marshaler interface.
// If ExplicitNull is true, returns "null".
// If Value is set, returns the value as JSON.
// If neither, returns nil (field is omitted).
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
// When the JSON value is "null", sets ExplicitNull to true.
// Otherwise, unmarshals the value into Value.
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

// LoadFromTerraform loads a value from Terraform into a NullableInt.
// If the field is not set in Terraform, sets Value to nil and ExplicitNull to false.
// If the field is set to -1, sets Value to nil and ExplicitNull to true.
// Otherwise, sets Value to the field's value and ExplicitNull to false.
func (n *NullableInt) LoadFromTerraform(value interface{}) {
	if value == nil {
		n.Value = nil
		n.ExplicitNull = false
		return
	}
	v := value.(int)
	if v == -1 {
		n.Value = nil
		n.ExplicitNull = true
	} else {
		n.Value = &v
		n.ExplicitNull = false
	}
}

// ToTerraform converts a NullableInt to a Terraform value.
// If ExplicitNull is true, returns -1 (field was set to -1).
// If Value is nil and ExplicitNull is false, returns nil (field is unset).
// Otherwise, returns the value.
func (n *NullableInt) ToTerraform() interface{} {
	if n.ExplicitNull {
		return -1
	}
	if n.Value == nil {
		return nil
	}
	return *n.Value
}

// NewNullableIntFromTerraform returns a pointer to a NullableInt based on the Terraform value.
// Returns nil if the field is not set.
func NewNullableIntFromTerraform(value interface{}) *NullableInt {
	if value == nil {
		return nil
	}
	ni := &NullableInt{}
	ni.LoadFromTerraform(value)
	return ni
}

// NullableIntFromResource returns a pointer to a NullableInt from a ResourceData key, or nil if not set.
func NullableIntFromResource(d interface {
	GetOk(string) (interface{}, bool)
}, key string) *NullableInt {
	if v, ok := d.GetOk(key); ok {
		return NewNullableIntFromTerraform(v)
	}
	return nil
}
