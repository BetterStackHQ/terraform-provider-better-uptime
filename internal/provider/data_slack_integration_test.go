package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSlackIntegrationSkipsNullChannel(t *testing.T) {
	// A Slack integration with a null channel name must be skipped, not panic the lookup.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}
		if r.Method == http.MethodGet && r.RequestURI == "/api/v2/slack-integrations?page=1" {
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"slack_channel_name":null,"slack_team_name":"Team1"}},{"id":"2","attributes":{"slack_channel_name":"#target","slack_team_name":"Team2"}}],"pagination":{"next":null}}`))
			return
		}
		t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) { return New(WithURL(server.URL)), nil },
		},
		Steps: []resource.TestStep{{
			Config: `
			provider "betteruptime" {
				api_token = "foo"
			}
			data "betteruptime_slack_integration" "this" {
				slack_channel_name = "#target"
			}
			`,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "id", "2"),
				resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "slack_channel_name", "#target"),
			),
		}},
	})
}

func TestDataSlackIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
			t.Fail()
		}

		prefix := "/api/v2/slack-integrations"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"slack_team_id":"T123456","slack_team_name":"Team1","slack_channel_id":"C123456","slack_channel_name":"#channel1","slack_status":"active","integration_type":"verbose","on_call_notifications":true}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"slack_team_id":"T456789","slack_team_name":"Team2","slack_channel_id":"C456789","slack_channel_name":"#channel2","slack_status":"active","integration_type":"verbose","on_call_notifications":true,"notify_alongside_primary_responder":false,"notify_on_resolve":true,"channel_display_name":"incidents","channel_invitees":[{"type":"user","id":2}],"alert_options":[{"type":"team","id":1,"name":"Team","call":false,"sms":false,"email":true,"push":false,"critical_alert":false},{"type":"policy","id":3,"name":"Policy","team_id":1,"team_name":"Team"}],"status_page_update_resources":[{"type":"status_page","id":4}]}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
			t.Fail()
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
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "notify_alongside_primary_responder", "false"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "notify_on_resolve", "true"),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "channel_display_name", "incidents"),
					resource.TestCheckTypeSetElemNestedAttrs("data.betteruptime_slack_integration.this", "channel_invitee.*", map[string]string{"type": "user", "id": "2"}),
					resource.TestCheckResourceAttr("data.betteruptime_slack_integration.this", "alert_option.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.betteruptime_slack_integration.this", "alert_option.*", map[string]string{"type": "team", "id": "1", "email": "true"}),
					resource.TestCheckTypeSetElemNestedAttrs("data.betteruptime_slack_integration.this", "alert_option.*", map[string]string{"type": "policy", "id": "3"}),
					resource.TestCheckTypeSetElemNestedAttrs("data.betteruptime_slack_integration.this", "status_page_update_resource.*", map[string]string{"type": "status_page", "id": "4"}),
				),
			},
		},
	})
}
