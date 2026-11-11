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

func TestCalculateValidation(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string // ว่าง = ไม่เช็กข้อความ
	}{
		{"negative totalIncome", `{"totalIncome":-250000,"wht":0,"allowances":[]}`,
			400, `{"error":"totalIncome must be a positive number"}`},
		{"wht greater than totalIncome", `{"totalIncome":300000,"wht":400000,"allowances":[]}`,
			400, `{"error":"wht cannot be greater than totalIncome"}`},
		{"negative wht", `{"totalIncome":300000,"wht":-1,"allowances":[]}`,
			400, `{"error":"wht must be a non-negative number"}`},
		{"missing totalIncome", `{"wht":0,"allowances":[]}`,
			400, `{"error":"totalIncome is required"}`},
		{"totalIncome is a string", `{"totalIncome":"abc","wht":0,"allowances":[]}`, 400, ""},
		{"allowances is not an array", `{"totalIncome":300000,"wht":0,"allowances":"x"}`, 400, ""},
		{"malformed json", `{bad`, 400, ""},
		{"empty body", ``, 400, ""},
		// สองเคสนี้ต้อง "ผ่าน" (ไม่ใช่ 400)
		{"zero income is allowed", `{"totalIncome":0,"wht":0,"allowances":[]}`, 200, `{"tax":0}`},
		{"wht equal to income is allowed", `{"totalIncome":100000,"wht":100000,"allowances":[]}`, 200, `{"tax":-100000}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			got := strings.TrimSpace(rec.Body.String())
			if tc.wantBody != "" && got != tc.wantBody {
				t.Errorf("body = %q, want %q", got, tc.wantBody)
			}
			if tc.wantStatus == 400 {
				if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}
				if !strings.HasPrefix(got, `{"error":"`) {
					t.Errorf("body should be an error JSON, got %q", got)
				}
			}
		})
	}
}
