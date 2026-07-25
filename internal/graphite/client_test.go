package graphite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFetchMetricsIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics/index.json" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`["stats.foo.bar","stats.foo.baz"]`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	names, err := c.FetchMetricsIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"stats.foo.bar", "stats.foo.baz"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Errorf("got %v, want %v", names, want)
	}
}

func TestFetchMetricsIndexError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.FetchMetricsIndex(context.Background()); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestFetchRender(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"target":"stats.foo","datapoints":[[1.5,1000],[null,1060]]}]`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	series, err := c.FetchRender(context.Background(), RenderQuery{
		Targets:       []string{"stats.foo"},
		From:          "-1h",
		MaxDataPoints: 300,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(series) != 1 {
		t.Fatalf("got %d series, want 1", len(series))
	}
	if series[0].Target != "stats.foo" {
		t.Errorf("target = %q, want stats.foo", series[0].Target)
	}
	if len(series[0].Datapoints) != 2 {
		t.Fatalf("got %d datapoints, want 2", len(series[0].Datapoints))
	}
	if series[0].Datapoints[0].Value == nil || *series[0].Datapoints[0].Value != 1.5 {
		t.Errorf("datapoint 0 value = %v, want 1.5", series[0].Datapoints[0].Value)
	}
	if series[0].Datapoints[1].Value != nil {
		t.Errorf("datapoint 1 value = %v, want nil", *series[0].Datapoints[1].Value)
	}
	if gotQuery == "" {
		t.Error("expected non-empty query string")
	}
}

func TestFetchRenderUnnamedTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"target":"","datapoints":[]}]`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	series, err := c.FetchRender(context.Background(), RenderQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if series[0].Target != "[unnamed_target]" {
		t.Errorf("target = %q, want [unnamed_target]", series[0].Target)
	}
}

func TestBasicAuthFromURL(t *testing.T) {
	var gotUser, gotPass string
	var hadAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, hadAuth = r.BasicAuth()
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	u := srv.URL[len("http://"):]
	c, err := NewClient("http://alice:s3cret@" + u)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.FetchMetricsIndex(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !hadAuth || gotUser != "alice" || gotPass != "s3cret" {
		t.Errorf("got auth (%v, %q, %q), want (true, alice, s3cret)", hadAuth, gotUser, gotPass)
	}
}

func TestSetAuthOverridesURL(t *testing.T) {
	var gotUser, gotPass string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	c.SetAuth("bob", "hunter2")
	if _, err := c.FetchMetricsIndex(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotUser != "bob" || gotPass != "hunter2" {
		t.Errorf("got (%q, %q), want (bob, hunter2)", gotUser, gotPass)
	}
}

func TestRenderURLOmitsCredentials(t *testing.T) {
	c, err := NewClient("http://alice:s3cret@graphite.example.com")
	if err != nil {
		t.Fatal(err)
	}
	url := c.RenderURL(RenderQuery{Targets: []string{"stats.foo"}, From: "-1h"})
	if strings.Contains(url, "alice") || strings.Contains(url, "s3cret") {
		t.Errorf("RenderURL leaked credentials: %s", url)
	}
	if !strings.Contains(url, "/render") {
		t.Errorf("RenderURL missing /render path: %s", url)
	}
}

func TestFetchRenderImage(t *testing.T) {
	fakePNG := []byte("\x89PNG\r\n\x1a\nfake-png-data")
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(fakePNG)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	png, err := c.FetchRenderImage(context.Background(), ImageQuery{
		Targets:  []string{"stats.foo"},
		From:     "-1h",
		Width:    800,
		Height:   400,
		AreaMode: "stacked",
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(png) != string(fakePNG) {
		t.Errorf("got %q, want %q", png, fakePNG)
	}

	if got := gotQuery.Get("format"); got != "png" {
		t.Errorf("format = %q, want png", got)
	}
	if got := gotQuery.Get("areaMode"); got != "stacked" {
		t.Errorf("areaMode = %q, want stacked", got)
	}
	if got := gotQuery.Get("width"); got != "800" {
		t.Errorf("width = %q, want 800", got)
	}
	if got := gotQuery.Get("height"); got != "400" {
		t.Errorf("height = %q, want 400", got)
	}
	if got := gotQuery.Get("target"); got != "stats.foo" {
		t.Errorf("target = %q, want stats.foo", got)
	}
}

func TestFetchRenderImageError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.FetchRenderImage(context.Background(), ImageQuery{}); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}
