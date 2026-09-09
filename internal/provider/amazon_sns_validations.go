package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The envelope fields uptime accepts, from Integrations::IncomingWebhookTargetable::SNS_ENVELOPE_FIELDS.
var snsEnvelopeTargetFields = []string{"TopicArn", "MessageId", "Subject"}

// validateSnsEnvelopeTargetField rejects an sns_envelope target without a valid target_field.
// uptime validates presence and inclusion on the field and rule models, so without this the
// practitioner gets a 422 at apply time for something knowable at plan time. This is the same
// reason validateIntegrationRuleConditions exists.
func validateSnsEnvelopeTargetField(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	config := diff.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}

	isCreate := diff.Id() == ""

	fieldKeys := append(append([]string{}, FieldAttributes...), FieldsAttributes...)
	for _, key := range fieldKeys {
		if err := checkSnsEnvelopeBlocks(config, key, "field_target", isCreate); err != nil {
			return err
		}
	}
	for _, key := range RulesAttributes {
		if err := checkSnsEnvelopeBlocks(config, key, "rule_target", isCreate); err != nil {
			return err
		}
	}
	return nil
}

func checkSnsEnvelopeBlocks(config cty.Value, key, targetAttr string, isCreate bool) error {
	blocks := config.GetAttr(key)
	if blocks.IsNull() || !blocks.IsKnown() || blocks.LengthInt() == 0 {
		return nil
	}

	for i, block := range blocks.AsValueSlice() {
		if block.IsNull() || !block.IsKnown() {
			continue
		}
		target := block.GetAttr(targetAttr)
		if target.IsNull() || !target.IsKnown() || target.AsString() != "sns_envelope" {
			continue
		}

		targetField := block.GetAttr("target_field")
		// A value the practitioner interpolated from something else is not knowable while
		// planning. Refusing it would reject a config that is perfectly valid once applied.
		if !targetField.IsKnown() {
			continue
		}
		// target_field is Optional and Computed, so an absent one on an existing resource keeps
		// the value already stored, which may well be a valid envelope field. Only creation can
		// be sure that absent really means absent.
		if targetField.IsNull() || targetField.AsString() == "" {
			if isCreate {
				return fmt.Errorf("%s.%d: %s = \"sns_envelope\" requires target_field to be one of %s",
					key, i, targetAttr, strings.Join(snsEnvelopeTargetFields, ", "))
			}
			continue
		}
		if !containsString(snsEnvelopeTargetFields, targetField.AsString()) {
			return fmt.Errorf("%s.%d: target_field must be one of %s when %s = \"sns_envelope\", got %q",
				key, i, strings.Join(snsEnvelopeTargetFields, ", "), targetAttr, targetField.AsString())
		}
	}
	return nil
}
