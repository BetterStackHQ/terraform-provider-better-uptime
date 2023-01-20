package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
)

var integrationFieldSchema = map[string]*schema.Schema{
	"name": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"special_type": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"field_target": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"target_field": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"match_type": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"content": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"content_before": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"content_after": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
}

var integrationRuleSchema = map[string]*schema.Schema{
	"rule_target": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"target_field": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"match_type": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"content": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
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
var FieldAttributes = []string{"cause_field", "started_alert_id_field", "acknowledged_alert_id_field", "resolved_alert_id_field"}
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
