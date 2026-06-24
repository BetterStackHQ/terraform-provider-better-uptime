package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceTeamMember(t *testing.T) {
	var created bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodPost && r.RequestURI == "/api/v2/team-members":
			created = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"message":"Invitation sent","data":{"id":null,"type":"team_member_invitation","attributes":{"email":"test@example.com","first_name":null,"last_name":null,"invited_at":"2026-01-01T00:00:00Z","role":"responder","mobile_app_platforms":[]}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			if !created {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":"Team member with email \"test@example.com\" not found"}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":null,"type":"team_member_invitation","attributes":{"email":"test@example.com","first_name":null,"last_name":null,"invited_at":"2026-01-01T00:00:00Z","role":"responder","mobile_app_platforms":[]}}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			created = false
			w.WriteHeader(http.StatusNoContent)
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

				resource "betteruptime_team_member" "this" {
					email = "test@example.com"
					role  = "responder"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "id", "test@example.com"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "email", "test@example.com"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "role", "responder"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "invited_at", "2026-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "member_id", ""),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "mobile_app_platforms.#", "0"),
				),
			},
			// No changes — plan should be empty.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_team_member" "this" {
					email = "test@example.com"
					role  = "responder"
				}
				`,
				PlanOnly: true,
			},
		},
	})
}

func TestResourceTeamMemberExistingUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodPost && r.RequestURI == "/api/v2/team-members":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":"User is already a member of the team","data":{"id":"42","type":"team_member","attributes":{"email":"existing@example.com","first_name":"Jane","last_name":"Doe","created_at":"2025-06-15T10:30:00Z","role":"admin","mobile_app_platforms":["ios"]}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			_, _ = w.Write([]byte(`{"data":{"id":"42","type":"team_member","attributes":{"email":"existing@example.com","first_name":"Jane","last_name":"Doe","created_at":"2025-06-15T10:30:00Z","role":"admin","mobile_app_platforms":["ios"]}}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			w.WriteHeader(http.StatusNoContent)
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

				resource "betteruptime_team_member" "this" {
					email = "existing@example.com"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "id", "existing@example.com"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "email", "existing@example.com"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "role", "admin"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "first_name", "Jane"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "last_name", "Doe"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "member_id", "42"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "created_at", "2025-06-15T10:30:00Z"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "mobile_app_platforms.#", "1"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "mobile_app_platforms.0", "ios"),
				),
			},
		},
	})
}

func TestResourceTeamMemberRoleUpdate(t *testing.T) {
	var changeRoleCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/team-members":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"7","type":"team_member","attributes":{"email":"x@example.com","role":"responder","role_id":11}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/team-members":
			role := "responder"
			if changeRoleCalled {
				role = "team_lead"
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":"7","type":"team_member","attributes":{"email":"x@example.com","role":"%s","role_id":11}}}`, role)))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/roles":
			_, _ = w.Write([]byte(`{"data":[{"id":"99","type":"role","attributes":{"name":"Team lead","role":"team_lead"}}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/team-members/7/change-role/99":
			changeRoleCalled = true
			_, _ = w.Write([]byte(`{"data":{"id":"7","type":"team_member","attributes":{"email":"x@example.com","role":"team_lead","role_id":99}}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/team-members":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: map[string]func() (*schema.Provider, error){"betteruptime": func() (*schema.Provider, error) { return New(WithURL(server.URL)), nil }},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}
				resource "betteruptime_team_member" "this" {
					email = "x@example.com"
					role  = "responder"
				}`,
				Check: resource.TestCheckResourceAttr("betteruptime_team_member.this", "role", "responder"),
			},
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}
				resource "betteruptime_team_member" "this" {
					email = "x@example.com"
					role  = "team_lead"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "role", "team_lead"),
					func(*terraform.State) error {
						if !changeRoleCalled {
							return fmt.Errorf("expected change-role endpoint to be called")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestResourceTeamMemberCustomRoleNoDiff(t *testing.T) {
	var changeRoleCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/team-members":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"5","type":"team_member","attributes":{"email":"custom@example.com","role":"custom","role_id":77}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/team-members":
			_, _ = w.Write([]byte(`{"data":{"id":"5","type":"team_member","attributes":{"email":"custom@example.com","role":"custom","role_id":77}}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/team-members":
			w.WriteHeader(http.StatusNoContent)
		case strings.Contains(r.URL.Path, "/change-role/"):
			changeRoleCalled = true
			t.Errorf("change-role endpoint must not be called for a custom-role member")
			w.WriteHeader(http.StatusBadRequest)
		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: map[string]func() (*schema.Provider, error){"betteruptime": func() (*schema.Provider, error) { return New(WithURL(server.URL)), nil }},
		Steps: []resource.TestStep{
			{
				// Config has a system role; API returns "custom". DiffSuppressFunc must
				// prevent any perpetual diff and must not trigger a change-role call.
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}
				resource "betteruptime_team_member" "this" {
					email = "custom@example.com"
					role  = "member"
				}`,
				Check: func(*terraform.State) error {
					if changeRoleCalled {
						return fmt.Errorf("change-role endpoint was called but must not be for a custom-role member")
					}
					return nil
				},
			},
		},
	})
}

func TestResourceTeamMemberImport(t *testing.T) {
	var created bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodPost && r.RequestURI == "/api/v2/team-members":
			created = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"message":"Invitation sent","data":{"id":null,"type":"team_member_invitation","attributes":{"email":"import@example.com","first_name":null,"last_name":null,"invited_at":"2026-02-01T00:00:00Z","role":"member","mobile_app_platforms":[]}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":null,"type":"team_member_invitation","attributes":{"email":"import@example.com","first_name":null,"last_name":null,"invited_at":"2026-02-01T00:00:00Z","role":"member","mobile_app_platforms":[]}}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			created = false
			w.WriteHeader(http.StatusNoContent)
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

				resource "betteruptime_team_member" "this" {
					email = "import@example.com"
					role  = "member"
				}
				`,
			},
			{
				ResourceName:      "betteruptime_team_member.this",
				ImportState:       true,
				ImportStateId:     "import@example.com",
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceTeamMemberCustomRoleByID(t *testing.T) {
	var createBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/team-members":
			b, _ := io.ReadAll(r.Body)
			createBody = string(b)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"7","type":"team_member","attributes":{"email":"byid@example.com","role":"custom","role_id":206,"mobile_app_platforms":[]}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/team-members":
			_, _ = w.Write([]byte(`{"data":{"id":"7","type":"team_member","attributes":{"email":"byid@example.com","role":"custom","role_id":206,"mobile_app_platforms":[]}}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/team-members":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: map[string]func() (*schema.Provider, error){"betteruptime": func() (*schema.Provider, error) { return New(WithURL(server.URL)), nil }},
		Steps: []resource.TestStep{
			{
				// Assigning a custom role by id must round-trip without a perpetual diff
				// (role comes back as "custom", role_id matches the configured value).
				Config: `
				provider "betteruptime" { api_token = "foo" }
				resource "betteruptime_team_member" "this" {
					email   = "byid@example.com"
					role_id = "206"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "role_id", "206"),
					resource.TestCheckResourceAttr("betteruptime_team_member.this", "role", "custom"),
					func(*terraform.State) error {
						if !strings.Contains(createBody, `"role_id":206`) {
							return fmt.Errorf("create body should send role_id, got: %s", createBody)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestResourceTeamMemberRoleConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: map[string]func() (*schema.Provider, error){"betteruptime": func() (*schema.Provider, error) { return New(WithURL("https://betterstack.com")), nil }},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "betteruptime" { api_token = "foo" }
				resource "betteruptime_team_member" "this" {
					email   = "x@example.com"
					role    = "member"
					role_id = "206"
				}`,
				ExpectError: regexp.MustCompile(`(?i)only one of`),
			},
		},
	})
}
