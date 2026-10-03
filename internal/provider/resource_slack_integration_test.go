package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func slackIntegrationConfig(body string) string {
	return testProviderBlock + `
	resource "betteruptime_slack_integration" "this" {
		better_stack_id = "184"
	` + body + `
	}
	`
}

const slackIntegrationAdoptedBody = `
		integration_type                      = "verbose"
		private_channel                       = false
		channel_display_name                  = "incidents"
		alert_all_teams_and_policies          = false
		invite_new_slack_users                = true
		on_call_notifications                 = false
		notify_on_pause                       = false
		notify_on_resolve                     = true
		direct_message_on_mention             = false
		ai_sre_on_mention                     = false
		restrict_incident_creation_to_channel = true
		post_timeline_events                  = true
		post_incident_metadata                = false
		post_status_page_updates              = true
		notify_alongside_primary_responder    = false

		channel_invitee {
			type = "user"
			id   = 2
		}
		channel_invitee {
			type = "on_call_calendar"
			id   = 1
		}
		alert_option {
			type  = "team"
			id    = 1
			email = true
			push  = true
		}
		alert_option {
			type = "policy"
			id   = 1
		}
		status_page_update_resource {
			type = "status_page"
			id   = 1
		}
`

// Every attribute differs from slackIntegrationAdoptedBody, bar post_status_page_updates (which
// turns status page updates off, so the last step flips it while clearing the lists).
func slackIntegrationChangedBody(notifyOnPause, lists string) string {
	return fmt.Sprintf(`
		integration_type                      = "thread"
		private_channel                       = true
		alert_all_teams_and_policies          = true
		invite_new_slack_users                = false
		on_call_notifications                 = true
		notify_on_pause                       = %s
		notify_on_resolve                     = false
		direct_message_on_mention             = true
		ai_sre_on_mention                     = true
		restrict_incident_creation_to_channel = false
		post_timeline_events                  = false
		post_incident_metadata                = true
		notify_alongside_primary_responder    = true
	%s
	`, notifyOnPause, lists)
}

const slackIntegrationChangedLists = `
		channel_display_name     = "war-room"
		post_status_page_updates = true

		channel_invitee {
			type = "slack_user_group"
			id   = 5
		}
		alert_option {
			type           = "team"
			id             = 2
			call           = true
			sms            = true
			critical_alert = true
		}
		status_page_update_resource {
			type = "monitor"
			id   = 7
		}
`

func (ts *TestServer) testCheckLastRequestBodyContains(method string, needles ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		var last *CalledRequest
		for i := range ts.CalledRequests {
			if ts.CalledRequests[i].Method == method {
				last = &ts.CalledRequests[i]
			}
		}
		if last == nil {
			return fmt.Errorf("no %s request was made", method)
		}
		for _, needle := range needles {
			if !strings.Contains(last.Body, needle) {
				return fmt.Errorf("expected last %s body to contain %s, got %s", method, needle, last.Body)
			}
		}
		return nil
	}
}

func TestResourceSlackIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/slack-integrations", "184")
	defer server.Close()
	// Connected in the web UI, since the API can't create a Slack integration.
	server.Data.Store([]byte(`{
		"slack_team_id": "TU9442",
		"slack_team_name": "U-9442 workspace",
		"slack_channel_id": "CU9442",
		"slack_channel_name": "#u9442-incidents",
		"slack_status": "active",
		"integration_type": "legacy",
		"channel_display_name": null,
		"channel_invitees": [],
		"private_channel": false,
		"alert_all_teams_and_policies": true,
		"alert_options": [],
		"invite_new_slack_users": false,
		"on_call_notifications": true,
		"notify_on_pause": true,
		"notify_on_resolve": false,
		"direct_message_on_mention": true,
		"ai_sre_on_mention": true,
		"restrict_incident_creation_to_channel": false,
		"post_timeline_events": false,
		"post_incident_metadata": true,
		"post_status_page_updates": false,
		"status_page_update_resources": [],
		"notify_alongside_primary_responder": true,
		"team_name": "Better Uptime"
	}`))

	const name = "betteruptime_slack_integration.this"
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Adopt by better_stack_id: every configured attribute is sent, false booleans included.
			{
				Config: slackIntegrationConfig(slackIntegrationAdoptedBody),
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("GET", "/api/v2/slack-integrations/184", ""),
					server.TestCheckCalledRequestCount("POST", "/api/v2/slack-integrations", 0),
					server.testCheckLastRequestBodyContains("PATCH",
						`"integration_type":"verbose"`,
						`"private_channel":false`,
						`"channel_display_name":"incidents"`,
						`"alert_all_teams_and_policies":false`,
						`"invite_new_slack_users":true`,
						`"on_call_notifications":false`,
						`"notify_on_pause":false`,
						`"notify_on_resolve":true`,
						`"direct_message_on_mention":false`,
						`"ai_sre_on_mention":false`,
						`"restrict_incident_creation_to_channel":true`,
						`"post_timeline_events":true`,
						`"post_incident_metadata":false`,
						`"post_status_page_updates":true`,
						`"notify_alongside_primary_responder":false`,
						`{"type":"user","id":2}`,
						`{"type":"on_call_calendar","id":1}`,
						// A team gets every alert channel, an escalation policy none (the API refuses them).
						`{"type":"team","id":1,"call":false,"sms":false,"email":true,"push":true,"critical_alert":false}`,
						`{"type":"policy","id":1}`,
						`"status_page_update_resources":[{"type":"status_page","id":1}]`,
					),
					resource.TestCheckResourceAttr(name, "id", "184"),
					resource.TestCheckResourceAttr(name, "better_stack_id", "184"),
					resource.TestCheckResourceAttr(name, "slack_channel_name", "#u9442-incidents"),
					resource.TestCheckResourceAttr(name, "slack_team_name", "U-9442 workspace"),
					resource.TestCheckResourceAttr(name, "integration_type", "verbose"),
					resource.TestCheckResourceAttr(name, "notify_on_pause", "false"),
					resource.TestCheckResourceAttr(name, "channel_invitee.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "channel_invitee.*", map[string]string{"type": "on_call_calendar", "id": "1"}),
					resource.TestCheckResourceAttr(name, "alert_option.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "alert_option.*", map[string]string{"type": "team", "id": "1", "email": "true", "push": "true"}),
					resource.TestCheckTypeSetElemNestedAttrs(name, "alert_option.*", map[string]string{"type": "policy", "id": "1"}),
					resource.TestCheckResourceAttr(name, "status_page_update_resource.#", "1"),
				),
			},
			// Change every attribute: only changed keys are sent, lists replace the whole list.
			{
				Config: slackIntegrationConfig(slackIntegrationChangedBody("true", slackIntegrationChangedLists)),
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("PATCH", "/api/v2/slack-integrations/184",
						`{"integration_type":"thread","channel_invitees":[{"type":"slack_user_group","id":5}],"private_channel":true,"channel_display_name":"war-room","alert_all_teams_and_policies":true,`+
							`"alert_options":[{"type":"team","id":2,"call":true,"sms":true,"email":false,"push":false,"critical_alert":true}],"invite_new_slack_users":false,"on_call_notifications":true,"notify_on_pause":true,`+
							`"notify_on_resolve":false,"direct_message_on_mention":true,"ai_sre_on_mention":true,"restrict_incident_creation_to_channel":false,"post_timeline_events":false,"post_incident_metadata":true,`+
							`"status_page_update_resources":[{"type":"monitor","id":7}],"notify_alongside_primary_responder":true}`),
					resource.TestCheckResourceAttr(name, "integration_type", "thread"),
					resource.TestCheckResourceAttr(name, "channel_display_name", "war-room"),
					resource.TestCheckResourceAttr(name, "ai_sre_on_mention", "true"),
					resource.TestCheckResourceAttr(name, "channel_invitee.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "channel_invitee.*", map[string]string{"type": "slack_user_group", "id": "5"}),
					resource.TestCheckResourceAttr(name, "alert_option.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "alert_option.*", map[string]string{"type": "team", "id": "2", "call": "true", "sms": "true", "critical_alert": "true"}),
					resource.TestCheckResourceAttr(name, "status_page_update_resource.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "status_page_update_resource.*", map[string]string{"type": "monitor", "id": "7"}),
				),
			},
			// A single change sends a single key, false included.
			{
				Config: slackIntegrationConfig(slackIntegrationChangedBody("false", slackIntegrationChangedLists)),
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("PATCH", "/api/v2/slack-integrations/184", `{"notify_on_pause":false}`),
					resource.TestCheckResourceAttr(name, "notify_on_pause", "false"),
				),
			},
			// Removing every block clears the list, and removing channel_display_name clears it.
			{
				Config: slackIntegrationConfig(slackIntegrationChangedBody("false", `
		post_status_page_updates = false
				`)),
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("PATCH", "/api/v2/slack-integrations/184",
						`{"channel_invitees":[],"channel_display_name":"","alert_options":[],"post_status_page_updates":false,"status_page_update_resources":[]}`),
					resource.TestCheckResourceAttr(name, "channel_display_name", ""),
					resource.TestCheckResourceAttr(name, "post_status_page_updates", "false"),
					resource.TestCheckResourceAttr(name, "channel_invitee.#", "0"),
					resource.TestCheckResourceAttr(name, "alert_option.#", "0"),
					resource.TestCheckResourceAttr(name, "status_page_update_resource.#", "0"),
				),
			},
			// Lists come back, so import has something to read.
			{
				Config: slackIntegrationConfig(slackIntegrationChangedBody("false", slackIntegrationChangedLists)),
				Check:  resource.TestCheckResourceAttr(name, "alert_option.#", "1"),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     "184",
				ImportStateVerify: true,
			},
		},
	})

	// Destroying only forgets the integration.
	if count := len(server.CalledRequests); count == 0 {
		t.Fatal("no requests were made")
	}
	for _, req := range server.CalledRequests {
		if req.Method == "DELETE" {
			t.Fatalf("destroy must not delete the Slack integration, got %s %s", req.Method, req.URL)
		}
	}
}

func TestResourceSlackIntegrationImport(t *testing.T) {
	server := newResourceServer(t, "/api/v2/slack-integrations", "184")
	defer server.Close()
	server.Data.Store([]byte(`{
		"slack_channel_name": "#u9442-incidents",
		"integration_type": "channel",
		"private_channel": true,
		"notify_on_pause": false,
		"channel_invitees": [{"type": "user", "id": 2}],
		"alert_options": [
			{"type": "team", "id": 1, "name": "Team", "call": false, "sms": false, "email": true, "push": false, "critical_alert": false},
			{"type": "policy", "id": 3, "name": "Policy", "team_id": 1, "team_name": "Team"}
		],
		"status_page_update_resources": [{"type": "heartbeat", "id": 9}]
	}`))

	const name = "betteruptime_slack_integration.this"
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Import a config-free integration, then manage it with a config that matches: no changes.
			{
				Config: slackIntegrationConfig(`
		channel_invitee {
			type = "user"
			id   = 2
		}
		alert_option {
			type  = "team"
			id    = 1
			email = true
		}
		alert_option {
			type = "policy"
			id   = 3
		}
		status_page_update_resource {
			type = "heartbeat"
			id   = 9
		}
				`),
				ResourceName:       name,
				ImportState:        true,
				ImportStateId:      "184",
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}
					attrs := states[0].Attributes
					for k, want := range map[string]string{
						"id":                            "184",
						"better_stack_id":               "184",
						"integration_type":              "channel",
						"private_channel":               "true",
						"channel_invitee.#":             "1",
						"alert_option.#":                "2",
						"status_page_update_resource.#": "1",
					} {
						if attrs[k] != want {
							return fmt.Errorf("expected %s = %q, got %q", k, want, attrs[k])
						}
					}
					return nil
				},
			},
			{
				Config: slackIntegrationConfig(`
		channel_invitee {
			type = "user"
			id   = 2
		}
		alert_option {
			type  = "team"
			id    = 1
			email = true
		}
		alert_option {
			type = "policy"
			id   = 3
		}
		status_page_update_resource {
			type = "heartbeat"
			id   = 9
		}
				`),
				PlanOnly: true,
			},
		},
	})
	if count := len(server.CalledRequests); count == 0 {
		t.Fatal("no requests were made")
	}
	for _, req := range server.CalledRequests {
		if req.Method == "PATCH" {
			t.Fatalf("import and an unchanged plan must not update, got %s %s %s", req.Method, req.URL, req.Body)
		}
	}
}

func TestResourceSlackIntegrationWithoutId(t *testing.T) {
	server := newResourceServer(t, "/api/v2/slack-integrations", "184")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: testProviderBlock + `
				resource "betteruptime_slack_integration" "this" {
					integration_type = "verbose"
				}
				`,
				ExpectError: regexp.MustCompile(`the integration has to be connected and removed in Better Stack web UI`),
			},
		},
	})
}

func TestResourceSlackIntegrationRejectsLegacyType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories("http://127.0.0.1:1"),
		Steps: []resource.TestStep{
			{
				Config:      slackIntegrationConfig(`integration_type = "legacy"`),
				ExpectError: regexp.MustCompile(`expected integration_type to be one of`),
			},
		},
	})
}

func TestResourceSlackIntegrationAPIError(t *testing.T) {
	server := newResourceServer(t, "/api/v2/slack-integrations", "184")
	defer server.Close()
	server.Data.Store([]byte(`{"slack_channel_name": "#u9442-incidents", "integration_type": "verbose"}`))
	server.ExpectRequest("PATCH", "/api/v2/slack-integrations/184", "", 422, `{"errors":{"alert_options":["Team 8 was not found."]}}`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: slackIntegrationConfig(`
		alert_option {
			type = "team"
			id   = 8
		}
				`),
				ExpectError: regexp.MustCompile(`Team 8 was not found`),
			},
		},
	})
}

func TestResourceSlackIntegrationAdoptConverges(t *testing.T) {
	server := newResourceServer(t, "/api/v2/slack-integrations", "184")
	defer server.Close()
	server.Data.Store([]byte(`{
		"slack_channel_name": "#u9442-incidents",
		"integration_type": "thread",
		"notify_on_pause": true,
		"channel_display_name": "incidents",
		"channel_invitees": [{"type": "user", "id": 2}],
		"alert_options": [{"type": "policy", "id": 3}],
		"status_page_update_resources": [{"type": "heartbeat", "id": 9}]
	}`))

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// What the config leaves out is kept where the API fills it in, and cleared where it doesn't,
			// so the plan after the adopting apply is empty (checked by the test framework).
			{
				Config: slackIntegrationConfig(`notify_on_pause = false`),
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("PATCH", "/api/v2/slack-integrations/184",
						`{"channel_invitees":[],"channel_display_name":"","alert_options":[],"notify_on_pause":false,"status_page_update_resources":[]}`),
					resource.TestCheckResourceAttr("betteruptime_slack_integration.this", "integration_type", "thread"),
				),
			},
		},
	})
}
