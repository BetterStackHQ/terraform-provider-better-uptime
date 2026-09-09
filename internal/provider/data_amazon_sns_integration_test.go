package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataAmazonSnsIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		prefix := "/api/v2/amazon-sns"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"name":"Amazon SNS 1","url":"https://uptime.betterstack.com/api/v1/incoming-webhook/abc","topic_arn":null,"subscription_state":"awaiting"}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"name":"Amazon SNS 2","url":"https://uptime.betterstack.com/api/v1/incoming-webhook/def","topic_arn":"arn:aws:sns:eu-central-1:123456789012:backend-alerts","subscription_state":"active"}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: testProviderBlock + `
				data "betteruptime_amazon_sns_integration" "this" {
					name = "Amazon SNS 2"
				}

				data "betteruptime_amazon_sns_integration" "awaiting" {
					name = "Amazon SNS 1"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.this", "id", "2"),
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.this", "name", "Amazon SNS 2"),
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.this", "topic_arn", "arn:aws:sns:eu-central-1:123456789012:backend-alerts"),
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.this", "subscription_state", "active"),
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.this", "url", "https://uptime.betterstack.com/api/v1/incoming-webhook/def"),
					// An integration whose subscription AWS has not confirmed yet has no topic_arn at
					// all. The lookup must still return it rather than skipping or panicking on the null.
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.awaiting", "id", "1"),
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.awaiting", "subscription_state", "awaiting"),
					resource.TestCheckResourceAttr("data.betteruptime_amazon_sns_integration.awaiting", "topic_arn", ""),
				),
			},
		},
	})
}
