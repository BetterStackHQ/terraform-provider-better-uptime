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

	fieldKeys := append(append([]string{}, FieldAttributes...), FieldsAttributes...)
	for _, key := range fieldKeys {
		if err := checkSnsEnvelopeBlocks(config, key, "field_target"); err != nil {
			return err
		}
	}
	for _, key := range RulesAttributes {
		if err := checkSnsEnvelopeBlocks(config, key, "rule_target"); err != nil {
			return err
		}
	}
	return nil
}

func checkSnsEnvelopeBlocks(config cty.Value, key, targetAttr string) error {
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
		if targetField.IsNull() || !targetField.IsKnown() || targetField.AsString() == "" {
			return fmt.Errorf("%s.%d: %s = \"sns_envelope\" requires target_field to be one of %s",
				key, i, targetAttr, strings.Join(snsEnvelopeTargetFields, ", "))
		}
		if !containsString(snsEnvelopeTargetFields, targetField.AsString()) {
			return fmt.Errorf("%s.%d: target_field must be one of %s when %s = \"sns_envelope\", got %q",
				key, i, strings.Join(snsEnvelopeTargetFields, ", "), targetAttr, targetField.AsString())
		}
	}
	return nil
}

// validateAmazonSnsTitleFieldNotRemoved rejects deleting a title_field block. title_field is
// Computed on this resource, because the API fills it in with the envelope's Subject when the
// integration is created, so removing the block from configuration produces no diff and Terraform
// reports "No changes" while the extraction stays in place. Erroring is the same choice
// validateTeamNameNotChanged makes: a silent no-op is worse than a refusal.
//
// Read from raw state and raw config rather than HasChange, which is false precisely because the
// attribute is Computed and absent from configuration.
func validateAmazonSnsTitleFieldNotRemoved(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" {
		return nil
	}

	state := diff.GetRawState()
	config := diff.GetRawConfig()
	if state.IsNull() || !state.IsKnown() || config.IsNull() || !config.IsKnown() {
		return nil
	}

	stateField := state.GetAttr("title_field")
	if stateField.IsNull() || !stateField.IsKnown() || stateField.LengthInt() == 0 {
		return nil
	}

	configField := config.GetAttr("title_field")
	if configField.IsNull() || (configField.IsKnown() && configField.LengthInt() == 0) {
		return fmt.Errorf("title_field cannot be removed through Terraform, because the API treats an " +
			"absent title_field as \"keep the current one\". Change the block in place instead, or remove " +
			"the extraction in the Better Stack UI")
	}
	return nil
}
