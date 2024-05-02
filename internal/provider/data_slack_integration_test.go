package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSlackIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		prefix := "/api/v2/slack-integrations"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"slack_team_id":"T123456","slack_team_name":"Team1","slack_channel_id":"C123456","slack_channel_name":"#channel1","slack_status":"active","integration_type":"verbose","on_call_notifications":true}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"slack_team_id":"T456789","slack_team_name":"Team2","slack_channel_id":"C456789","slack_channel_name":"#channel2","slack_status":"active","integration_type":"verbose","on_call_notifications":true}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))
	defer server.Close()

	var slackChannelName = "#channel2"

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

				data "betteruptime_slack_integration" "this" {
					slack_channel_name = "%s"
				}
				`, slackChannelName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_slack_integration.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "slack_team_id", "T456789"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "slack_team_name", "Team2"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "slack_channel_id", "C456789"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "slack_channel_name", slackChannelName),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "slack_status", "active"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "integration_type", "verbose"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "on_call_notifications", "true"),
				),
			},
		},
	})
}
