package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Description: "The target of the field. Can be any of the following: from_email, subject, or body for email integrations or query_string, header, body, json and xml for incoming webhooks.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"target_field": {
		Description: "The target field within the content of the field_target. Should be a JSON key when field_target is json, a CSS selector when field_target is XML, name of the header for headers or a parameter name for query parameters",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"match_type": {
		Description: "The match type of the field. Can be any of the following: match_before, match_after, match_between, match_regex, or match_everything.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"content": {
		Description: "How should we extract content the field. Should be a valid Regex when match_type is match_regex.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"content_before": {
		Description: "When should we stop extracting content for the field. Should be present when match_type is either match_between or match_before.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"content_after": {
		Description: "When should we start extracting content for the field. Should be present when match_type is either match_between or match_after.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
}

var integrationRuleSchema = map[string]*schema.Schema{
	"rule_target": {
		Description: "The target of the rule. Can be any of the following: from_email, subject, or body for email integrations or query_string, header, body, json and xml for incoming webhooks.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"target_field": {
		Description: "The target field within the content of the rule_target. Should be a JSON key when rule_target is json, a CSS selector when rule_target is XML, name of the header for headers or a parameter name for query parameters",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"match_type": {
		Description: "The type of the rule. Can be any of the following: contains, contains_not, matches_regex, matches_regex_not, equals, or equals_not.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"content": {
		Description: "The content we should match to satisfy the rule. Should be a valid Regex when match_type is match_regex.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
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
	fieldValues := d.Get(key).(*schema.Set).List()
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
