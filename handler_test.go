package liveflux

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/hb"
)

type handlerComp struct {
	Base
	Count int
}

func TestHandler_GetScript(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/liveflux", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/javascript") {
		t.Fatalf("expected javascript content-type, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "window.liveflux") {
		t.Fatalf("expected script body to contain window.liveflux, got %q", body)
	}
}

func (c *handlerComp) GetKind() string { return "" }
func (c *handlerComp) Mount(_ context.Context, params map[string]string) error {
	if v, ok := params["init"]; ok && v == "err" {
		return errors.New("mount-fail")
	}
	if v, ok := params["init"]; ok && v != "" {
		// naive parse to int
		if v == "1" {
			c.Count = 1
		}
	}
	return nil
}
func (c *handlerComp) Handle(_ context.Context, action string, _ url.Values) error {
	if action == "inc" {
		c.Count++
		return nil
	}
	if action == "redir" {
		c.Redirect("/next", 2)
		return nil
	}
	return errors.New("bad-action")
}
func (c *handlerComp) Render(_ context.Context) hb.TagInterface {
	return hb.Div().Attr("data-id", c.GetID()).Text(fmt.Sprintf("count=%d", c.Count))
}

// helper to register a fresh kind each test
func registerTestKind(t *testing.T, proto ComponentInterface) string {
	t.Helper()
	kind := proto.GetKind()
	if kind == "" {
		kind = fmt.Sprintf("test-comp-%p", proto)
	}
	_ = RegisterByKind(kind, proto)
	return kind
}

// targetComp implements TargetRenderer for testing targeted updates
type targetComp struct {
	Base
	Value string
	dirty map[string]bool
}

func (c *targetComp) GetKind() string { return "target-test" }

func (c *targetComp) Mount(_ context.Context, params map[string]string) error {
	c.Value = params["value"]
	c.dirty = make(map[string]bool)
	return nil
}

func (c *targetComp) Handle(_ context.Context, action string, data url.Values) error {
	if action == "update" {
		c.Value = data.Get("value")
		c.dirty["result"] = true
	}
	return nil
}

func (c *targetComp) Render(_ context.Context) hb.TagInterface {
	return hb.Div().
		Attr("data-id", c.GetID()).
		Children([]hb.TagInterface{
			hb.Input().Name("value").Value(c.Value),
			hb.Div().ID("result").Text("Result: " + c.Value),
		})
}

func (c *targetComp) RenderTargets(_ context.Context) []TargetFragment {
	var targets []TargetFragment
	if c.dirty["result"] {
		targets = append(targets, TargetFragment{
			Selector: "#result",
			Content:  hb.Div().ID("result").Text("Result: " + c.Value),
		})
	}
	return targets
}

func TestHandler_TargetedRendering(t *testing.T) {
	s := NewMemoryStore()
	h := NewHandler(s)
	kind := registerTestKind(t, &targetComp{})

	// Mount component
	mountForm := url.Values{FormComponentKind: {kind}, "value": {"initial"}}
	mountReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(mountForm.Encode()))
	mountReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mountRec := httptest.NewRecorder()
	h.ServeHTTP(mountRec, mountReq)

	html := mountRec.Body.String()
	start := strings.Index(html, "data-id=\"")
	if start < 0 {
		t.Fatalf("no data-id in mount HTML: %s", html)
	}
	start += len("data-id=\"")
	end := strings.Index(html[start:], "\"")
	id := html[start : start+end]

	// Trigger action that marks target dirty
	actForm := url.Values{
		FormComponentKind: {kind},
		FormComponentID:   {id},
		FormAction:        {"update"},
		"value":           {"updated"},
	}
	actReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(actForm.Encode()))
	actReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	actRec := httptest.NewRecorder()
	h.ServeHTTP(actRec, actReq)

	if actRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on action, got %d", actRec.Code)
	}

	body := actRec.Body.String()

	// Should contain targeted template, not full component replacement
	if !strings.Contains(body, `<template data-flux-target="#result"`) {
		t.Errorf("expected targeted template, got: %s", body)
	}

	// Should contain updated value in fragment
	if !strings.Contains(body, "Result: updated") {
		t.Errorf("expected updated value in fragment, got: %s", body)
	}

	// Should NOT contain full component fallback template
	if strings.Contains(body, `<template data-flux-component-kind=`) {
		t.Errorf("should not contain full component fallback template, got: %s", body)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Result().StatusCode)
	}
}

func TestHandler_MountMissingKind(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "missing component kind") {
		t.Fatalf("expected 400 missing component kind, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestHandler_MountNotRegistered(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	form := url.Values{FormComponentKind: {"unknown.kind"}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unregistered component, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestHandler_MountSuccess_RendersAndStores(t *testing.T) {
	s := NewMemoryStore()
	h := NewHandler(s)
	kind := registerTestKind(t, &handlerComp{})
	form := url.Values{
		FormComponentKind: {kind},
		"init":            {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	html := rec.Body.String()
	if !strings.Contains(html, "count=1") {
		t.Fatalf("expected rendered HTML to contain count=1, got: %s", html)
	}
	// Extract id attribute
	if !strings.Contains(html, "data-id=\"") {
		t.Fatalf("expected data-id attribute in HTML")
	}
}

func TestHandler_MountError(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	kind := registerTestKind(t, &handlerComp{})
	form := url.Values{
		FormComponentKind: {kind},
		"init":            {"err"},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "mount error") {
		t.Fatalf("expected 500 mount error, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestHandler_HandleValidateMissing(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	// Provide a non-empty id to take the handle path, but leave kind empty to trigger validation error
	form := url.Values{FormComponentKind: {""}, FormComponentID: {"some-id"}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "missing component or id") {
		t.Fatalf("expected 400 missing component or id, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestHandler_HandleActionSuccess(t *testing.T) {
	s := NewMemoryStore()
	h := NewHandler(s)
	kind := registerTestKind(t, &handlerComp{})
	// First mount to create and store
	mountForm := url.Values{FormComponentKind: {kind}}
	mountReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(mountForm.Encode()))
	mountReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mountRec := httptest.NewRecorder()
	h.ServeHTTP(mountRec, mountReq)
	html := mountRec.Body.String()
	// find id="..." in data-id
	start := strings.Index(html, "data-id=\"")
	if start < 0 {
		t.Fatalf("no data-id in mount HTML: %s", html)
	}
	start += len("data-id=\"")
	end := strings.Index(html[start:], "\"")
	id := html[start : start+end]

	// act inc
	actForm := url.Values{FormComponentKind: {kind}, FormComponentID: {id}, FormAction: {"inc"}}
	actReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(actForm.Encode()))
	actReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	actRec := httptest.NewRecorder()
	h.ServeHTTP(actRec, actReq)
	if actRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on action, got %d", actRec.Code)
	}
	if !strings.Contains(actRec.Body.String(), "count=1") {
		t.Fatalf("expected updated render with count=1, got: %s", actRec.Body.String())
	}
}

func TestHandler_HandleActionError(t *testing.T) {
	s := NewMemoryStore()
	h := NewHandler(s)
	kind := registerTestKind(t, &handlerComp{})
	// mount
	mountForm := url.Values{FormComponentKind: {kind}}
	mountReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(mountForm.Encode()))
	mountReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mountRec := httptest.NewRecorder()
	h.ServeHTTP(mountRec, mountReq)
	html := mountRec.Body.String()
	start := strings.Index(html, "data-id=\"")
	start += len("data-id=\"")
	end := strings.Index(html[start:], "\"")
	id := html[start : start+end]

	// act invalid
	actForm := url.Values{FormComponentKind: {kind}, FormComponentID: {id}, FormAction: {"oops"}}
	actReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(actForm.Encode()))
	actReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	actRec := httptest.NewRecorder()
	h.ServeHTTP(actRec, actReq)
	if actRec.Code != http.StatusBadRequest || !strings.Contains(actRec.Body.String(), "action error") {
		t.Fatalf("expected 400 action error, got %d %q", actRec.Code, actRec.Body.String())
	}
}

func TestHandler_HandleRedirect(t *testing.T) {
	s := NewMemoryStore()
	h := NewHandler(s)
	kind := registerTestKind(t, &handlerComp{})
	// mount
	mountForm := url.Values{FormComponentKind: {kind}}
	mountReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(mountForm.Encode()))
	mountReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mountRec := httptest.NewRecorder()
	h.ServeHTTP(mountRec, mountReq)
	html := mountRec.Body.String()
	start := strings.Index(html, "data-id=\"")
	start += len("data-id=\"")
	end := strings.Index(html[start:], "\"")
	id := html[start : start+end]

	// act request redirect
	actForm := url.Values{FormComponentKind: {kind}, FormComponentID: {id}, FormAction: {"redir"}}
	actReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(actForm.Encode()))
	actReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	actRec := httptest.NewRecorder()
	h.ServeHTTP(actRec, actReq)

	if actRec.Code != http.StatusOK { // body still 200 with redirect HTML
		t.Fatalf("expected 200 with redirect fallback body, got %d", actRec.Code)
	}
	hdr := actRec.Result().Header
	if hdr.Get("X-Liveflux-Redirect") != "/next" {
		t.Fatalf("missing X-Liveflux-Redirect header, got %q", hdr.Get("X-Liveflux-Redirect"))
	}
	if hdr.Get("X-Liveflux-Redirect-After") != "2" {
		t.Fatalf("missing/incorrect redirect delay header, got %q", hdr.Get("X-Liveflux-Redirect-After"))
	}
	body := actRec.Body.String()
	if !strings.Contains(body, "<noscript><meta http-equiv=\"refresh\"") || !strings.Contains(body, "window.location.replace") {
		t.Fatalf("expected redirect fallback HTML, got: %s", body)
	}
}
