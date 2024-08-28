package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestIpListData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		switch {
		case r.Method == http.MethodGet && r.RequestURI == "/ips-by-cluster.json":
			_, _ = w.Write([]byte(`{"eu":["139.162.215.1","139.162.215.2"],"us":["66.228.56.1","66.228.56.2","66.228.56.3"]}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
			t.Fail()
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

				data "betteruptime_ip_list" "this" {
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_ip_list.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.0", "139.162.215.1"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.1", "139.162.215.2"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.2", "66.228.56.1"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.3", "66.228.56.2"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.4", "66.228.56.3"),
					resource.TestCheckNoResourceAttr("data.betteruptime_ip_list.this", "ips.5"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "all_clusters.0", "eu"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "all_clusters.1", "us"),
					resource.TestCheckNoResourceAttr("data.betteruptime_ip_list.this", "all_clusters.2"),
				),
			},
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				data "betteruptime_ip_list" "this" {
					filter_clusters = ["us"]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_ip_list.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.0", "66.228.56.1"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.1", "66.228.56.2"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "ips.2", "66.228.56.3"),
					resource.TestCheckNoResourceAttr("data.betteruptime_ip_list.this", "ips.3"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "all_clusters.0", "eu"),
					resource.TestCheckResourceAttr("data.betteruptime_ip_list.this", "all_clusters.1", "us"),
					resource.TestCheckNoResourceAttr("data.betteruptime_ip_list.this", "all_clusters.2"),
				),
			},
		},
	})
}
