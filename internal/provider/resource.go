package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func resourceCreate(ctx context.Context, meta interface{}, url string, in, out interface{}) diag.Diagnostics {
	reqBody, err := json.Marshal(&in)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("POST %s: %s", url, string(reqBody))
	res, err := meta.(*client).Post(ctx, url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		// Keep-Alive.
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return diag.Errorf("POST %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	}
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("POST %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	if err := json.Unmarshal(body, &out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRead(ctx context.Context, meta interface{}, url string, out interface{}) (derr diag.Diagnostics, ok bool) {
	log.Printf("GET %s", url)
	res, err := meta.(*client).Get(ctx, url)
	if err != nil {
		return diag.FromErr(err), false
	}
	defer func() {
		// Keep-Alive.
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	if res.StatusCode == http.StatusNotFound {
		return nil, false
	}
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return diag.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body)), false
	}
	if err != nil {
		return diag.FromErr(err), false
	}
	log.Printf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	err = json.Unmarshal(body, &out)
	if err != nil {
		return diag.FromErr(err), false
	}
	return nil, true
}

func resourceUpdate(ctx context.Context, meta interface{}, url string, req interface{}, out interface{}) diag.Diagnostics {
	reqBody, err := json.Marshal(&req)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("PATCH %s: %s", url, string(reqBody))
	res, err := meta.(*client).Patch(ctx, url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		// Keep-Alive.
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return diag.Errorf("PATCH %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	}
	log.Printf("PATCH %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	if err := json.Unmarshal(body, &out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
func resourceDelete(ctx context.Context, meta interface{}, url string) diag.Diagnostics {
	log.Printf("DELETE %s", url)
	res, err := meta.(*client).Delete(ctx, url)
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		// Keep-Alive.
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusNotFound {
		return diag.Errorf("DELETE %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	}
	log.Printf("DELETE %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	return nil
}
