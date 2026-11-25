package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestBoundarySetsNoStoreOnErrors(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = safeEchoError
	e.Use(boundaryMiddleware())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/missing", nil))
	if rec.Code != http.StatusNotFound || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected response: status=%d cache=%q", rec.Code, rec.Header().Get("Cache-Control"))
	}
}
