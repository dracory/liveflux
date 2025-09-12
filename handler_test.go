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

func (c *handlerComp) GetAlias() string { return "" }
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

// helper to register a fresh alias each test
func registerTestAlias(t *testing.T, proto ComponentInterface) string {
	alias := "test." + NewID()
	RegisterByAlias(alias, proto)
	return alias
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

func TestHandler_MountMissingAlias(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "missing component alias") {
		t.Fatalf("expected 400 missing component alias, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestHandler_MountNotRegistered(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	form := url.Values{FormComponent: {"unknown.alias"}}
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
	alias := registerTestAlias(t, &handlerComp{})
	form := url.Values{
		FormComponent: {alias},
		"init":       {"1"},
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
	alias := registerTestAlias(t, &handlerComp{})
	form := url.Values{
		FormComponent: {alias},
		"init":       {"err"},
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
	// Provide a non-empty id to take the handle path, but leave alias empty to trigger validation error
	form := url.Values{FormComponent: {""}, FormID: {"some-id"}}
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
	alias := registerTestAlias(t, &handlerComp{})
	// First mount to create and store
	mountForm := url.Values{FormComponent: {alias}}
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
	actForm := url.Values{FormComponent: {alias}, FormID: {id}, FormAction: {"inc"}}
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
	alias := registerTestAlias(t, &handlerComp{})
	// mount
	mountForm := url.Values{FormComponent: {alias}}
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
	actForm := url.Values{FormComponent: {alias}, FormID: {id}, FormAction: {"oops"}}
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
	alias := registerTestAlias(t, &handlerComp{})
	// mount
	mountForm := url.Values{FormComponent: {alias}}
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
	actForm := url.Values{FormComponent: {alias}, FormID: {id}, FormAction: {"redir"}}
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
