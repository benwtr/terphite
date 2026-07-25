// Package graphite is a small HTTP client for the parts of the Graphite web
// API terphite needs: the metrics index and the render endpoint.
package graphite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client talks to a single Graphite server.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient builds a Client for the given Graphite base URL, e.g.
// "http://user:pass@graphite.example.com:8080". Credentials embedded in the
// URL are used for HTTP Basic Auth; call SetAuth to override them.
func NewClient(rawURL string) (*Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("graphite: invalid URL %q: %w", rawURL, err)
	}
	return &Client{BaseURL: u, HTTPClient: http.DefaultClient}, nil
}

// SetAuth sets (or overrides) the HTTP Basic Auth credentials used for
// requests, taking precedence over any userinfo embedded in the base URL.
func (c *Client) SetAuth(username, password string) {
	u := *c.BaseURL
	u.User = url.UserPassword(username, password)
	c.BaseURL = &u
}

func (c *Client) newRequest(ctx context.Context, path string, query url.Values) (*http.Request, error) {
	u := *c.BaseURL
	user := u.User
	u.User = nil
	u.Path = strings.TrimRight(u.Path, "/") + path
	if query != nil {
		u.RawQuery = query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if user != nil {
		pass, _ := user.Password()
		req.SetBasicAuth(user.Username(), pass)
	}
	return req, nil
}

// FetchMetricsIndex returns the flat list of all metric names known to the
// server, from GET /metrics/index.json.
func (c *Client) FetchMetricsIndex(ctx context.Context) ([]string, error) {
	req, err := c.newRequest(ctx, "/metrics/index.json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphite: fetching metrics index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphite: metrics index returned status %d", resp.StatusCode)
	}
	var names []string
	if err := json.NewDecoder(resp.Body).Decode(&names); err != nil {
		return nil, fmt.Errorf("graphite: decoding metrics index: %w", err)
	}
	return names, nil
}

// RenderQuery describes a request to the render endpoint.
type RenderQuery struct {
	Targets       []string
	From          string
	MaxDataPoints int
}

func (q RenderQuery) values() url.Values {
	v := url.Values{}
	v.Set("format", "json")
	if q.From != "" {
		v.Set("from", q.From)
	}
	for _, t := range q.Targets {
		v.Add("target", t)
	}
	if q.MaxDataPoints > 0 {
		v.Set("maxDataPoints", strconv.Itoa(q.MaxDataPoints))
	}
	return v
}

// RenderURL builds the render URL for q without embedding credentials, e.g.
// for "open in browser" or "copy to clipboard" actions.
func (c *Client) RenderURL(q RenderQuery) string {
	u := *c.BaseURL
	u.User = nil
	u.Path = strings.TrimRight(u.Path, "/") + "/render"
	u.RawQuery = q.values().Encode()
	return u.String()
}

// Datapoint is a single (possibly missing) sample from a render response.
type Datapoint struct {
	Value *float64
	Time  time.Time
}

// Series is one target's worth of render data.
type Series struct {
	Target     string
	Datapoints []Datapoint
}

// FetchRender fetches render data for q from GET /render.
func (c *Client) FetchRender(ctx context.Context, q RenderQuery) ([]Series, error) {
	req, err := c.newRequest(ctx, "/render", q.values())
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphite: fetching render data: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphite: render returned status %d", resp.StatusCode)
	}

	var raw []struct {
		Target     string        `json:"target"`
		Datapoints [][2]*float64 `json:"datapoints"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("graphite: decoding render data: %w", err)
	}

	series := make([]Series, 0, len(raw))
	for _, r := range raw {
		title := r.Target
		if title == "" {
			title = "[unnamed_target]"
		}
		dps := make([]Datapoint, 0, len(r.Datapoints))
		for _, pair := range r.Datapoints {
			if pair[1] == nil {
				continue
			}
			dp := Datapoint{Time: time.Unix(int64(*pair[1]), 0).UTC()}
			if pair[0] != nil {
				v := *pair[0]
				dp.Value = &v
			}
			dps = append(dps, dp)
		}
		series = append(series, Series{Target: title, Datapoints: dps})
	}
	return series, nil
}
