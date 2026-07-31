package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The API omits scenario_name and url for monitor types that don't have them;
// monitorCopyAttrs must not dereference the missing fields (issue #228: a refresh
// after a failed in-place type change panicked with a nil dereference).
func TestMonitorCopyAttrsNilOptionalFields(t *testing.T) {
	t.Run("scenario_name in state, absent in API response", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, monitorSchema, map[string]interface{}{
			"scenario_name": "httpbingo probe",
			"monitor_type":  "keyword",
		})
		monitorType := "keyword"
		monitorUrl := "https://example.com"
		if derr := monitorCopyAttrs(d, &monitor{MonitorType: &monitorType, URL: &monitorUrl}); derr != nil {
			t.Fatal(derr)
		}
		if got := d.Get("scenario_name").(string); got != "httpbingo probe" {
			t.Fatalf("scenario_name = %q, want the state value preserved", got)
		}
	})

	t.Run("url in state, absent in API response", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, monitorSchema, map[string]interface{}{
			"url":          "https://example.com",
			"monitor_type": "playwright",
		})
		monitorType := "playwright"
		if derr := monitorCopyAttrs(d, &monitor{MonitorType: &monitorType}); derr != nil {
			t.Fatal(derr)
		}
		if got := d.Get("url").(string); got != "https://example.com" {
			t.Fatalf("url = %q, want the state value preserved", got)
		}
	})
}

func TestResourceMonitorTypeChangeToPlaywright(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create a keyword monitor.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url              = "https://example.com"
					monitor_type     = "keyword"
					required_keyword = "example"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", "keyword"),
				),
			},
			// Step 2 - in-place changes within the non-playwright types stay allowed.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "https://example.com"
					monitor_type = "status"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", "status"),
				),
			},
			// Step 3 - changing to playwright must fail at plan time: the API rejects it
			// and the failed apply would corrupt state (issue #228).
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					monitor_type      = "playwright"
					scenario_name     = "example probe"
					playwright_script = "// script"
				}
				`,
				ExpectError: regexp.MustCompile(`monitor_type cannot be changed to or from 'playwright'`),
			},
		},
	})
}

func TestResourceMonitorTypeChangeFromPlaywright(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create a playwright monitor.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					monitor_type      = "playwright"
					scenario_name     = "example probe"
					playwright_script = "// script"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", "playwright"),
				),
			},
			// Step 2 - changing away from playwright must fail at plan time.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "https://example.com"
					monitor_type = "status"
				}
				`,
				ExpectError: regexp.MustCompile(`monitor_type cannot be changed to or from 'playwright'`),
			},
		},
	})
}
