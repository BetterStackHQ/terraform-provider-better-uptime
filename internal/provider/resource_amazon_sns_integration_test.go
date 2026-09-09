package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const amazonSnsProviderBlock = `
provider "betteruptime" {
  api_token = "foo"
}
`

func amazonSnsProviderFactories(server *TestServer) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"betteruptime": func() (*schema.Provider, error) {
			return New(WithURL(server.URL)), nil
		},
	}
}

// Applies the example config we actually ship, rather than a copy of it. The E2E job is the only
// other thing that exercises the examples, and it needs live credentials, so an example that does
// not even pass schema validation could reach the registry as a resource's headline documentation.
func TestResourceAmazonSnsIntegrationShippedExample(t *testing.T) {
	example, err := os.ReadFile("../../examples/resources/betteruptime_amazon_sns_integration/resource.tf")
	if err != nil {
		t.Fatal(err)
	}

	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: amazonSnsProviderBlock + string(example),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_amazon_sns_integration.backend_alerts", "id"),
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.backend_alerts", "name", "Terraform Amazon SNS"),
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.backend_alerts", "title_field.0.field_target", "sns_envelope"),
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.backend_alerts", "title_field.0.target_field", "Subject"),
				),
			},
		},
	})
}

// The envelope carries TopicArn, MessageId and Subject, which is the whole reason Amazon SNS needs
// its own target: they are not in the message body the other targets read.
func TestResourceAmazonSnsIntegrationSnsEnvelopeTarget(t *testing.T) {
	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				  started_rules {
					rule_target  = "sns_envelope"
					target_field = "TopicArn"
					match_type   = "contains"
					content      = "arn:aws:sns:eu-central-1:123456789012:backend-alerts"
				  }
				  title_field {
					field_target = "sns_envelope"
					target_field = "Subject"
					match_type   = "match_everything"
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.this", "started_rules.0.rule_target", "sns_envelope"),
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.this", "started_rules.0.target_field", "TopicArn"),
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.this", "title_field.0.field_target", "sns_envelope"),
				),
			},
		},
	})
}

// The API gives a fresh Amazon SNS integration a title extracted from the envelope's Subject, but
// it reads a present-and-null title_field as "destroy the title". Sending the key unconditionally,
// as the incoming webhook resource does, would therefore throw that default away on every create.
func TestResourceAmazonSnsIntegrationKeepsApiTitleDefault(t *testing.T) {
	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "unused"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				}`,
				Check: server.TestCheckCalledRequestWithout("POST", "/api/v2/amazon-sns", "title_field"),
			},
		},
	})
}

// An sns_envelope target without a valid target_field is a 422 at apply time, and the whole point
// of the resource's other CustomizeDiff checks is to catch that class of error while planning.
func TestResourceAmazonSnsIntegrationRejectsBadSnsEnvelopeTargetField(t *testing.T) {
	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "unused"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				  title_field {
					field_target = "sns_envelope"
					match_type   = "match_everything"
				  }
				}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`title_field\.0: field_target = "sns_envelope" requires target_field to be one of TopicArn, MessageId, Subject`),
			},
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				  started_rules {
					rule_target  = "sns_envelope"
					target_field = "Topic"
					match_type   = "contains"
					content      = "backend"
				  }
				}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`started_rules\.0: target_field must be one of TopicArn, MessageId, Subject when rule_target = "sns_envelope", got "Topic"`),
			},
		},
	})
}

// title_field is Computed, so deleting the block produces no diff at all. Terraform would say
// "No changes" while the extraction stayed in place, which is the silent no-op team_name already
// refuses to make.
func TestResourceAmazonSnsIntegrationRefusesToRemoveTitleField(t *testing.T) {
	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "unused"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				  title_field {
					field_target = "sns_envelope"
					target_field = "Subject"
					match_type   = "match_everything"
				  }
				}`,
			},
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "unused"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				}`,
				ExpectError: regexp.MustCompile(`title_field cannot be removed through Terraform`),
			},
		},
	})
}

// The update path sends only what changed. This is what makes an explicit read-only write guard
// unnecessary: a computed attribute with no configuration is never seen as changed, so topic_arn
// and subscription_state stay out of the request without anything filtering them. title_field must
// stay out too, since the API reads a present-and-null one as "destroy the title".
func TestResourceAmazonSnsIntegrationUpdateSendsOnlyChangedAttributes(t *testing.T) {
	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	config := func(name string) string {
		return amazonSnsProviderBlock + `
		resource "betteruptime_amazon_sns_integration" "this" {
		  name                   = "` + name + `"
		  started_rule_type      = "unused"
		  acknowledged_rule_type = "unused"
		  resolved_rule_type     = "unused"
		}`
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: config("Terraform Test"),
			},
			{
				Config: config("Terraform Test - Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.this", "name", "Terraform Test - Updated"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/amazon-sns/1", `{"name":"Terraform Test - Updated"}`),
					server.TestCheckCalledRequestWithout("PATCH", "/api/v2/amazon-sns/1", "topic_arn"),
					server.TestCheckCalledRequestWithout("PATCH", "/api/v2/amazon-sns/1", "subscription_state"),
					server.TestCheckCalledRequestWithout("PATCH", "/api/v2/amazon-sns/1", "title_field"),
				),
			},
			{
				ResourceName:      "betteruptime_amazon_sns_integration.this",
				ImportState:       true,
				ImportStateId:     "1",
				ImportStateVerify: false, // Cannot verify due to TypeSet fields not properly transformed
			},
		},
	})
}

// topic_arn and subscription_state are written by Better Stack when AWS answers the subscription
// handshake. They must arrive in state from the API and never be sent back as configuration.
func TestResourceAmazonSnsIntegrationReadOnlyAttributes(t *testing.T) {
	server := newResourceServer(t, "/api/v2/amazon-sns", "1")
	defer server.Close()

	const topicArn = "arn:aws:sns:eu-central-1:123456789012:backend-alerts"
	server.ExpectRequest("POST", "/api/v2/amazon-sns", "", 201,
		`{"data":{"id":"1","attributes":{"name":"Terraform Test","started_rule_type":"unused","acknowledged_rule_type":"unused","resolved_rule_type":"unused","topic_arn":"`+topicArn+`","subscription_state":"active"}}}`)
	server.ExpectRequest("GET", "/api/v2/amazon-sns/1", "", 200,
		`{"data":{"id":"1","attributes":{"name":"Terraform Test","started_rule_type":"unused","acknowledged_rule_type":"unused","resolved_rule_type":"unused","topic_arn":"`+topicArn+`","subscription_state":"active"}}}`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: amazonSnsProviderFactories(server),
		Steps: []resource.TestStep{
			{
				Config: amazonSnsProviderBlock + `
				resource "betteruptime_amazon_sns_integration" "this" {
				  name                   = "Terraform Test"
				  started_rule_type      = "unused"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type     = "unused"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.this", "topic_arn", topicArn),
					resource.TestCheckResourceAttr("betteruptime_amazon_sns_integration.this", "subscription_state", "active"),
					server.TestCheckCalledRequestWithout("POST", "/api/v2/amazon-sns", "topic_arn"),
					server.TestCheckCalledRequestWithout("POST", "/api/v2/amazon-sns", "subscription_state"),
				),
			},
		},
	})
}
