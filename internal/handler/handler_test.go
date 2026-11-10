package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string // ว่าง = ไม่เช็ก body
	}{
		{"health ok", http.MethodGet, "/health", "", http.StatusOK, `{"status":"ok"}`},
		{"health wrong method", http.MethodPost, "/health", "", http.StatusMethodNotAllowed, ""},
		{"unknown path", http.MethodGet, "/nope", "", http.StatusNotFound, ""},
		{"calc example 1", http.MethodPost, "/tax/calculations",
			`{"totalIncome":750000,"wht":0,"allowances":[]}`, http.StatusOK, `{"tax":63500}`},
		{"calc wrong method", http.MethodGet, "/tax/calculations", "", http.StatusMethodNotAllowed, ""},
		{"calc malformed json", http.MethodPost, "/tax/calculations", `{bad`, http.StatusBadRequest, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantBody != "" {
				got := strings.TrimSpace(rec.Body.String())
				if got != tc.wantBody {
					t.Errorf("body = %q, want %q", got, tc.wantBody)
				}
			}
		})
	}
}