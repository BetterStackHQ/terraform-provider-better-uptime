package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataIncomingWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		prefix := "/api/v2/incoming-webhooks"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"name": "Incoming Webhook 1", "url":"https://betteruptime.com/api/v1/incoming-webhook/abc","recovery_period":3, "team_wait":60}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"name": "Incoming Webhook 2", "url":"https://betteruptime.com/api/v1/incoming-webhook/def","recovery_period":4, "team_wait":120}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))
	defer server.Close()

	var name = "Incoming Webhook 2"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				data "betteruptime_incoming_webhook" "this" {
					name = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_incoming_webhook.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_incoming_webhook.this", "name", name),
					resource.TestCheckResourceAttr("data.betteruptime_incoming_webhook.this", "recovery_period", "4"),
					resource.TestCheckResourceAttr("data.betteruptime_incoming_webhook.this", "team_wait", "120"),
					resource.TestCheckResourceAttr("data.betteruptime_incoming_webhook.this", "url", "https://betteruptime.com/api/v1/incoming-webhook/def"),
				),
			},
		},
	})
}
