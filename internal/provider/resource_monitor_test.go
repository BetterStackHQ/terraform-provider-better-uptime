package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceMonitor(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					paused       = true
					regions      = ["us", "eu"]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "computed_by_betteruptime"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                = "%s"
					monitor_type       = "%s"
					pronounceable_name = "override"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "override"),
				),
			},
			// Step 3 - update (but preserve pronounceable_name).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					http_method  = "POST"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "http_method", "POST"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "override"),
				),
			},
			// Step 4 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					http_method  = "POST"
				}
				`, url, monitorType),
				PlanOnly: true,
			},
			// Step 5 - destroy.
			{
				ResourceName:      "betteruptime_monitor.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceMonitorWithHeaders(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST"
							value = "test"
						}
					]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "request_headers.0.id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.name", "X-TEST"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.value", "test"),
				),
			},
			// Step 2 - add another request header.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST"
							value = "test"
						},
						{
							name  = "X-TEST-2"
							value = "test-2"
						}
					]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.name", "X-TEST"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.value", "test"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.1.name", "X-TEST-2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.1.value", "test-2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.#", "2"),
				),
			},
			// Step 3 - remove the first header.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST-2"
							value = "test-2"
						}
					]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "request_headers.0.id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.name", "X-TEST-2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.value", "test-2"),
					resource.TestCheckNoResourceAttr("betteruptime_monitor.this", "request_headers.1.name"),
				),
			},
			// Step 4 - invalid header with empty name.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = ""
							value = "test"
						}
					]
				}
				`, url, monitorType),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid request header map\[name: value:test\]: must contain 'name' key with a non-empty string value`),
			},
			// Step 5 - invalid header with empty value.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST"
							value = ""
						}
					]
				}
				`, url, monitorType),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid request header map\[name:X-TEST value:\]: must contain 'value' key with a non-empty string value`),
			},
			// Step 6 - invalid header with extra keys.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST"
							value = "test"
							extra = "invalid"
						}
					]
				}
				`, url, monitorType),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid request header map\[extra:invalid name:X-TEST value:test\]: must only contain 'name' and 'value' keys`),
			},
			// Step 7 - invalid header with incorrect format.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							"X-TEST" = "test"
						}
					]
				}
				`, url, monitorType),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`invalid request header map\[X-TEST:test\]: must contain 'name' key with a non-empty string value`),
			},
		},
	})
}

func TestResourceMonitorWithExpirationPolicyId(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expiration_policy_id", "0"),
					server.TestCheckCalledRequest("POST", "/api/v2/monitors", `{"expiration_policy_id":null,"url":"http://example.com","monitor_type":"status","request_headers":null}`),
				),
			},
			// Step 2 - update (set to non-null value).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                  = "%s"
					monitor_type         = "%s"
					expiration_policy_id = 5
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expiration_policy_id", "5"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"expiration_policy_id":5}`),
				),
			},
			// Step 3 - update (set to null).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                  = "%s"
					monitor_type         = "%s"
					expiration_policy_id = null
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expiration_policy_id", "0"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"expiration_policy_id":null}`),
				),
			},
			// Step 4 - update (set to non-null value again).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                  = "%s"
					monitor_type         = "%s"
					expiration_policy_id = 6
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expiration_policy_id", "6"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"expiration_policy_id":6}`),
				),
			},
			// Step 5 - update (unset).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expiration_policy_id", "0"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"expiration_policy_id":null}`),
				),
			},
			// Step 5 - destroy.
			{
				ResourceName:      "betteruptime_monitor.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceMonitorWithVariables(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "playwright"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                   = "%s"
					monitor_type          = "%s"
					environment_variables = {
						"TEST_1" = "test-1"
						"TEST_2" = "test-2"
					}
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.%", "2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.TEST_1", "test-1"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.TEST_2", "test-2"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					environment_variables = {
						"TEST_1" = "test-1"
						"TEST_3" = "test-3"
					}
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.%", "2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.TEST_1", "test-1"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.TEST_3", "test-3"),
				),
			},
			// Step 3 - invalid variable (empty name).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					environment_variables = {
						"" = "test"
					}
				}
				`, url, monitorType),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Environment variable name cannot be empty`),
			},
			// Step 4 - invalid variable (empty value).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					environment_variables = {
						TEST = ""
					}
				}
				`, url, monitorType),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Environment variable value for key "TEST" cannot be empty`),
			},
			// Step 5 - delete.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					environment_variables = {}
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "environment_variables.%", "0"),
				),
			},
			// Step 6 - destroy.
			{
				ResourceName:      "betteruptime_monitor.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestExpectedStatusCodeMonitor(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "expected_status_code"
	var expectedStatusCodes = []int{500, 501, 502}

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

				resource "betteruptime_monitor" "this" {
					url                   = "%s"
					monitor_type          = "%s"
					expected_status_codes = [%d, %d, %d]
					paused                = true
					regions               = ["us", "eu"]
				}
				`, url, monitorType, expectedStatusCodes[0], expectedStatusCodes[1], expectedStatusCodes[2]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.#", strconv.Itoa(len(expectedStatusCodes))),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.0", strconv.Itoa(expectedStatusCodes[0])),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.1", strconv.Itoa(expectedStatusCodes[1])),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.2", strconv.Itoa(expectedStatusCodes[2])),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "computed_by_betteruptime"),
				),
			},
		},
	})
}

func TestResourceMonitorWithDomainExpiration(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					server.TestCheckCalledRequest("POST", "/api/v2/monitors", `{"expiration_policy_id":null,"url":"http://example.com","monitor_type":"status","request_headers":null}`),
				),
			},
			// Step 2 - update (set to non-null value).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url              = "%s"
					monitor_type     = "%s"
					domain_expiration = 7
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "domain_expiration", "7"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"domain_expiration":7,"expiration_policy_id":null}`),
				),
			},
			// Step 3 - update (set to null).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url              = "%s"
					monitor_type     = "%s"
					domain_expiration = -1
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "domain_expiration", "-1"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"domain_expiration":null,"expiration_policy_id":null}`),
				),
			},
			// Step 4 - update (set to non-null value again).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url              = "%s"
					monitor_type     = "%s"
					domain_expiration = 14
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "domain_expiration", "14"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"domain_expiration":14,"expiration_policy_id":null}`),
				),
			},
			// Step 5 - update (unset).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					// TODO: Should not plan anything
				),
			},
			// Step 6 - destroy.
			{
				ResourceName:      "betteruptime_monitor.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceMonitorWithSSLExpiration(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					server.TestCheckCalledRequest("POST", "/api/v2/monitors", `{"expiration_policy_id":null,"url":"http://example.com","monitor_type":"status","request_headers":null}`),
				),
			},
			// Step 2 - update (set to non-null value).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url           = "%s"
					monitor_type  = "%s"
					ssl_expiration = 7
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "ssl_expiration", "7"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"ssl_expiration":7,"expiration_policy_id":null}`),
				),
			},
			// Step 3 - update (set to null).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url           = "%s"
					monitor_type  = "%s"
					ssl_expiration = -1
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "ssl_expiration", "-1"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"ssl_expiration":null,"expiration_policy_id":null}`),
				),
			},
			// Step 4 - update (set to non-null value again).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url           = "%s"
					monitor_type  = "%s"
					ssl_expiration = 14
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "ssl_expiration", "14"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/monitors/1", `{"ssl_expiration":14,"expiration_policy_id":null}`),
				),
			},
			// Step 5 - update (unset).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					// TODO: Should not plan anything
				),
			},
			// Step 6 - destroy.
			{
				ResourceName:      "betteruptime_monitor.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func createTestServer(t *testing.T) *TestServer {
	ts := &TestServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
			t.Fail()
		}
		ts.mu.Lock()
		ts.CalledRequests = append(ts.CalledRequests, CalledRequest{Method: r.Method, URL: r.RequestURI, Body: string(body)})
		ts.mu.Unlock()

		t.Log("Received " + r.Method + " " + r.RequestURI + " " + string(body))

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
			t.Fail()
		}

		// Check for custom request handlers first
		ts.mu.Lock()
		for _, expected := range ts.ExpectedRequests {
			if expected.Method == r.Method && expected.URL == r.RequestURI {
				if expected.Body == "" || expected.Body == string(body) {
					w.WriteHeader(expected.StatusCode)
					_, _ = w.Write([]byte(expected.Response))
					ts.mu.Unlock()
					return
				}
			}
		}
		ts.mu.Unlock()

		prefix := "/api/v2/monitors"
		id := "1"

		switch {
		case r.Method == http.MethodPost && r.RequestURI == prefix:
			ts.Data.Store(body)
			w.WriteHeader(http.StatusCreated)
			// Inject pronounceable_name.
			computed := make(map[string]interface{})
			if err := json.Unmarshal(body, &computed); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			computed["pronounceable_name"] = "computed_by_betteruptime"
			if computed["request_headers"] != nil {
				computed["request_headers"] = processRequestHeaders(computed["request_headers"].([]interface{}))
			}
			body, err = json.Marshal(computed)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, body)))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"/"+id:
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, ts.Data.Load().([]byte))))
		case r.Method == http.MethodPatch && r.RequestURI == prefix+"/"+id:
			patch := make(map[string]interface{})
			if err = json.Unmarshal(ts.Data.Load().([]byte), &patch); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			if err = json.Unmarshal(body, &patch); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			if patch["request_headers"] != nil {
				patch["request_headers"] = processRequestHeaders(patch["request_headers"].([]interface{}))
			}
			patched, err := json.Marshal(patch)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			ts.Data.Store(patched)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, patched)))
		case r.Method == http.MethodDelete && r.RequestURI == prefix+"/"+id:
			w.WriteHeader(http.StatusNoContent)
			ts.Data.Store([]byte(nil))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
			t.Fail()
		}
	}))

	return ts
}

func processRequestHeaders(headers []interface{}) []interface{} {
	var newHeaders []interface{}

	for _, v := range headers {
		header := v.(map[string]interface{})
		header["id"] = "1"

		if _, ok := header["_destroy"]; !ok {
			newHeaders = append(newHeaders, header)
		}
	}

	return newHeaders
}
