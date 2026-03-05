package provider

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataTeamMember(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodGet && r.RequestURI == "/api/v2/team-members?email=petr%40betterstack.com":
			_, _ = w.Write([]byte(`{"data":{"id":"7","type":"team_member","attributes":{"email":"petr@betterstack.com","first_name":"Petr","last_name":"Heinz","created_at":"2024-01-15T08:00:00Z","role":"admin"}}}`))
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

				data "betteruptime_team_member" "this" {
					email = "petr@betterstack.com"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "id", "petr@betterstack.com"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "email", "petr@betterstack.com"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "first_name", "Petr"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "last_name", "Heinz"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "role", "admin"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "member_id", "7"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "created_at", "2024-01-15T08:00:00Z"),
				),
			},
		},
	})
}

func TestDataTeamMemberInvitation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodGet && r.RequestURI == "/api/v2/team-members?email=invited%40example.com":
			_, _ = w.Write([]byte(`{"data":{"id":null,"type":"team_member_invitation","attributes":{"email":"invited@example.com","first_name":null,"last_name":null,"invited_at":"2026-03-01T12:00:00Z","role":"responder"}}}`))
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

				data "betteruptime_team_member" "this" {
					email = "invited@example.com"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "id", "invited@example.com"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "email", "invited@example.com"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "role", "responder"),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "member_id", ""),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "first_name", ""),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "last_name", ""),
					resource.TestCheckResourceAttr("data.betteruptime_team_member.this", "invited_at", "2026-03-01T12:00:00Z"),
				),
			},
		},
	})
}

func TestDataTeamMemberNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodGet && r.RequestURI == "/api/v2/team-members?email=nobody%40example.com":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":"Team member with email \"nobody@example.com\" not found"}`))
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

				data "betteruptime_team_member" "this" {
					email = "nobody@example.com"
				}
				`,
				ExpectError: regexp.MustCompile(`team member with email "nobody@example.com" not found`),
			},
		},
	})
}
