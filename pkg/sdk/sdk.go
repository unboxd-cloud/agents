// Package sdk is the Go client for the Unboxd platform control plane. It reuses
// the same DTOs the services speak (internal/api and the domain types), so the
// SDK never duplicates the wire contract.
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/catalog"
	"github.com/unboxd-cloud/platform/internal/compliance"
	"github.com/unboxd-cloud/platform/internal/metering"
)

// Client talks to the four control-plane services. Zero value is not usable;
// use New.
type Client struct {
	Catalog    string
	Metering   string
	Billing    string
	Compliance string
	Tenant     string
	HTTP       *http.Client
}

// New returns a Client pointed at the standard local ports. Override the URL
// fields (and Tenant) as needed.
func New() *Client {
	return &Client{
		Catalog:    "http://localhost:8083",
		Metering:   "http://localhost:8081",
		Billing:    "http://localhost:8082",
		Compliance: "http://localhost:8084",
		HTTP:       &http.Client{Timeout: 10 * time.Second},
	}
}

// ListOfferings returns the catalog, optionally filtered by persona profile.
func (c *Client) ListOfferings(ctx context.Context, profile string) ([]catalog.Offering, error) {
	url := c.Catalog + "/v1/catalog"
	if profile != "" {
		url += "?profile=" + profile
	}
	var out []catalog.Offering
	return out, c.do(ctx, http.MethodGet, url, nil, &out)
}

// PriceBook returns the active price book.
func (c *Client) PriceBook(ctx context.Context) (billing.PriceBook, error) {
	var out billing.PriceBook
	return out, c.do(ctx, http.MethodGet, c.Billing+"/v1/pricebook", nil, &out)
}

// RecordUsage records a single usage event.
func (c *Client) RecordUsage(ctx context.Context, e metering.UsageEvent) error {
	return c.do(ctx, http.MethodPost, c.Metering+"/v1/usage", e, nil)
}

// Rate rates usage into an invoice (with optional settlement and tax).
func (c *Client) Rate(ctx context.Context, req api.RateRequest) (api.RateResponse, error) {
	var out api.RateResponse
	return out, c.do(ctx, http.MethodPost, c.Billing+"/v1/rate", req, &out)
}

// Frameworks returns the loaded compliance frameworks.
func (c *Client) Frameworks(ctx context.Context) ([]compliance.Spec, error) {
	var out []compliance.Spec
	return out, c.do(ctx, http.MethodGet, c.Compliance+"/v1/frameworks", nil, &out)
}

// Evaluate checks a placement against a tenant compliance profile.
func (c *Client) Evaluate(ctx context.Context, req api.EvalRequest) (compliance.Report, error) {
	var out compliance.Report
	return out, c.do(ctx, http.MethodPost, c.Compliance+"/v1/evaluate", req, &out)
}

func (c *Client) do(ctx context.Context, method, url string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Tenant != "" {
		req.Header.Set("X-Tenant-ID", c.Tenant)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Treat 2xx and the compliance 422 (non-compliant report) as decodable.
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusUnprocessableEntity {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: %s: %s", method, url, resp.Status, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
