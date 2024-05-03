package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestOnCallCalendarData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		prefix := "/api/v2/on-calls"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id":"123","attributes":{"name":"Primary","default_calendar":true},"relationships":{"on_call_users":{"data":[{"id":"123456","type":"user"}]}}}],"included":[{"id":"123456","type":"user","attributes":{"first_name":"John","last_name":"Smith","email":"john@example.com","phone_numbers":[]}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id":"456","attributes":{"name":"Secondary","default_calendar":false},"relationships":{"on_call_users":{"data":[{"id":"456789","type":"user"}]}}}],"included":[{"id":"456789","type":"user","attributes":{"first_name":"Jane","last_name":"Doe","email":"jane@example.com","phone_numbers":["+44 808 157 0192"]}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))
	defer server.Close()

	var calendarName = "Secondary"

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

				data "betteruptime_on_call_calendar" "this" {
					name = "%s"
				}
				`, calendarName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_on_call_calendar.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "name", calendarName),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "default_calendar", "false"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.id", "456789"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.first_name", "Jane"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.last_name", "Doe"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.email", "jane@example.com"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.phone_numbers.0", "+44 808 157 0192"),
				),
			},
		},
	})
}

func TestDefaultOnCallCalendarData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodGet && r.RequestURI == "/api/v2/on-calls/default":
			_, _ = w.Write([]byte(`{"data":{"id":"123","attributes":{"name":"Primary","default_calendar":true},"relationships":{"on_call_users":{"data":[{"id":"123456","type":"user"}]}}},"included":[{"id":"123456","type":"user","attributes":{"first_name":"John","last_name":"Smith","email":"john@example.com","phone_numbers":[]}}]}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				data "betteruptime_on_call_calendar" "this" {
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_on_call_calendar.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "name", "Primary"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "default_calendar", "true"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.id", "123456"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.first_name", "John"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.last_name", "Smith"),
					resource.TestCheckResourceAttr("data.betteruptime_on_call_calendar.this", "on_call_users.0.email", "john@example.com"),
				),
			},
		},
	})
}
