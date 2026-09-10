package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mitchellh/mapstructure"
)

var integrationFieldSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the field.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"special_type": {
		Description: "A special type of the field. Can be alert_id or cause or otherwise null for a custom field.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"field_target": {
		Description: "The target of the field.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"target_field": {
		Description: "The target field within the content of the field_target.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"match_type": {
		Description:  "The match type of the field. Can be any of the following: match_before, match_after, match_between, match_regex, or match_everything.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"match_before", "match_after", "match_between", "match_regex", "match_everything"}, false),
	},
	"content": {
		Description: "The content to match. Required when match_type is match_before, match_after, or match_regex. Should be a valid regular expression when match_type is match_regex.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"content_before": {
		Description: "When should we stop extracting content for the field. Required when match_type is match_between.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"content_after": {
		Description: "When should we start extracting content for the field. Required when match_type is match_between.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
}

var integrationRuleSchema = map[string]*schema.Schema{
	"rule_target": {
		Description: "The target of the rule.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"target_field": {
		Description: "The target field within the content of the rule_target.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"match_type": {
		Description:  "The type of the rule. Can be any of the following: contains, contains_not, matches_regex, matches_regex_not, equals, or equals_not.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"contains", "contains_not", "matches_regex", "matches_regex_not", "equals", "equals_not"}, false),
	},
	"content": {
		Description: "The content we should match to satisfy the rule. Should be a valid regular expression when match_type is matches_regex.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
}

// integrationTargets is what one integration type accepts as a field_target or rule_target. The
// shared field and rule schemas above are generic; each integration narrows a copy of them with
// integrationFieldSchemaFor and integrationRuleSchemaFor, so a target the type does not support
// fails while planning rather than as a 422 at apply, and its docs list only what applies.
type integrationTargets struct {
	// values is what field_target and rule_target accept.
	values []string
	// targetDescription follows the generic field_target and rule_target description.
	targetDescription string
	// targetFieldDescription, when set, follows the generic target_field description.
	targetFieldDescription string
}

func integrationFieldSchemaFor(targets integrationTargets) map[string]*schema.Schema {
	return narrowIntegrationSchema(integrationFieldSchema, "field_target", targets)
}

func integrationRuleSchemaFor(targets integrationTargets) map[string]*schema.Schema {
	return narrowIntegrationSchema(integrationRuleSchema, "rule_target", targets)
}

func narrowIntegrationSchema(base map[string]*schema.Schema, targetKey string, targets integrationTargets) map[string]*schema.Schema {
	s := make(map[string]*schema.Schema, len(base))
	for k, v := range base {
		cp := *v
		switch k {
		case targetKey:
			cp.Description += " " + targets.targetDescription
			cp.ValidateFunc = validation.StringInSlice(targets.values, false)
		case "target_field":
			if targets.targetFieldDescription != "" {
				cp.Description += " " + targets.targetFieldDescription
			}
		}
		s[k] = &cp
	}
	return s
}

type integrationField struct {
	Name          *string `mapstructure:"name,omitempty" json:"name,omitempty" structs:"name,omitempty"`
	SpecialType   *string `mapstructure:"special_type,omitempty" json:"special_type,omitempty" structs:"special_type,omitempty"`
	FieldTarget   *string `mapstructure:"field_target,omitempty" json:"field_target,omitempty" structs:"field_target,omitempty"`
	TargetField   *string `mapstructure:"target_field,omitempty" json:"target_field,omitempty" structs:"target_field,omitempty"`
	MatchType     *string `mapstructure:"match_type,omitempty" json:"match_type,omitempty" structs:"match_type,omitempty"`
	Content       *string `mapstructure:"content,omitempty" json:"content,omitempty" structs:"content,omitempty"`
	ContentBefore *string `mapstructure:"content_before,omitempty" json:"content_before,omitempty" structs:"content_before,omitempty"`
	ContentAfter  *string `mapstructure:"content_after,omitempty" json:"content_after,omitempty" structs:"content_after,omitempty"`
}

type integrationRule struct {
	RuleTarget  *string `mapstructure:"rule_target,omitempty" json:"rule_target,omitempty"`
	TargetField *string `mapstructure:"target_field,omitempty" json:"target_field,omitempty"`
	MatchType   *string `mapstructure:"match_type,omitempty" json:"match_type,omitempty"`
	Content     *string `mapstructure:"content,omitempty" json:"content,omitempty"`
}

var RulesAttributes = []string{"started_rules", "acknowledged_rules", "resolved_rules"}
var FieldAttributes = []string{"cause_field", "title_field", "started_alert_id_field", "acknowledged_alert_id_field", "resolved_alert_id_field"}
var FieldsAttributes = []string{"other_started_fields", "other_acknowledged_fields", "other_resolved_fields"}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func isRulesAttribute(attribute string) bool  { return containsString(RulesAttributes, attribute) }
func isFieldAttribute(attribute string) bool  { return containsString(FieldAttributes, attribute) }
func isFieldsAttribute(attribute string) bool { return containsString(FieldsAttributes, attribute) }

// "all" or "any" rule type requires at least 1 rule, otherwise the API call fails with 422
func validateIntegrationRuleConditions(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	config := diff.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	for _, pair := range []struct {
		typeKey  string
		rulesKey string
	}{
		{"started_rule_type", "started_rules"},
		{"acknowledged_rule_type", "acknowledged_rules"},
		{"resolved_rule_type", "resolved_rules"},
	} {
		typeAttr := config.GetAttr(pair.typeKey)
		if typeAttr.IsNull() || !typeAttr.IsKnown() {
			continue
		}
		ruleType := typeAttr.AsString()
		if ruleType != "all" && ruleType != "any" {
			continue
		}
		rulesAttr := config.GetAttr(pair.rulesKey)
		if rulesAttr.IsNull() || (rulesAttr.IsKnown() && rulesAttr.LengthInt() == 0) {
			return fmt.Errorf("%s = %q requires at least one %s block; use \"unused\" to match every message", pair.typeKey, ruleType, pair.rulesKey)
		}
	}
	return nil
}

func loadIntegrationRules(d *schema.ResourceData, key string, receiver **[]integrationRule) {
	x := receiver
	rulesValues := d.Get(key)

	var rules []integrationRule

	for _, ruleValues := range rulesValues.([]interface{}) {
		ruleValuesObject := ruleValues.(map[string]interface{})

		for k, v := range ruleValuesObject {
			if v == "" {
				ruleValuesObject[k] = nil
			}
		}

		var integrationRule integrationRule
		err := mapstructure.Decode(ruleValuesObject, &integrationRule)
		if err != nil {
			panic(err)
		}

		rules = append(rules, integrationRule)
	}

	if len(rules) > 0 {
		*x = &rules
	}
}

func loadIntegrationField(d *schema.ResourceData, key string, receiver **integrationField) {
	fieldValues := d.Get(key).([]interface{})
	if len(fieldValues) > 0 {
		x := receiver

		fieldValuesObject := fieldValues[0].(map[string]interface{})

		for k, v := range fieldValuesObject {
			if v == "" {
				fieldValuesObject[k] = nil
			}
		}

		var integrationField integrationField
		err := mapstructure.Decode(fieldValuesObject, &integrationField)
		if err != nil {
			panic(err)
		}

		*x = &integrationField
	}
}

func loadIntegrationFields(d *schema.ResourceData, key string, receiver **[]integrationField) {
	x := receiver
	fieldsValues := d.Get(key)

	var fields []integrationField

	for _, fieldValues := range fieldsValues.([]interface{}) {
		fieldValuesObject := fieldValues.(map[string]interface{})

		for k, v := range fieldValuesObject {
			if v == "" {
				fieldValuesObject[k] = nil
			}
		}

		var integrationField integrationField
		err := mapstructure.Decode(fieldValuesObject, &integrationField)
		if err != nil {
			panic(err)
		}

		fields = append(fields, integrationField)
	}

	if len(fields) > 0 {
		*x = &fields
	}
}
