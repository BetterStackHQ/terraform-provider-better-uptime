package provider

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			_, _ = w.Write([]byte(`{"message":"Invitation sent","data":{"id":null,"type":"team_member_invitation","attributes":{"email":"test@example.com","first_name":null,"last_name":null,"invited_at":"2026-01-01T00:00:00Z","role":"responder"}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			if !created {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":"Team member with email \"test@example.com\" not found"}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":null,"type":"team_member_invitation","attributes":{"email":"test@example.com","first_name":null,"last_name":null,"invited_at":"2026-01-01T00:00:00Z","role":"responder"}}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/api/v2/team-members/"):
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
			_, _ = w.Write([]byte(`{"message":"User is already a member of the team","data":{"id":"42","type":"team_member","attributes":{"email":"existing@example.com","first_name":"Jane","last_name":"Doe","created_at":"2025-06-15T10:30:00Z","role":"admin"}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			_, _ = w.Write([]byte(`{"data":{"id":"42","type":"team_member","attributes":{"email":"existing@example.com","first_name":"Jane","last_name":"Doe","created_at":"2025-06-15T10:30:00Z","role":"admin"}}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/api/v2/team-members/"):
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
				),
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
			_, _ = w.Write([]byte(`{"message":"Invitation sent","data":{"id":null,"type":"team_member_invitation","attributes":{"email":"import@example.com","first_name":null,"last_name":null,"invited_at":"2026-02-01T00:00:00Z","role":"member"}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.RequestURI, "/api/v2/team-members?email="):
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":null,"type":"team_member_invitation","attributes":{"email":"import@example.com","first_name":null,"last_name":null,"invited_at":"2026-02-01T00:00:00Z","role":"member"}}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/api/v2/team-members/"):
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
