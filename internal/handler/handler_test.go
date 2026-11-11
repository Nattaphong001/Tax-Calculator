package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ทดสอบเส้นทางและ HTTP method: /health, path ที่ไม่มี, method ผิด และ JSON ที่พัง
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
			`{"totalIncome":750000,"wht":0,"allowances":[]}`, http.StatusOK, ""},
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

// ทดสอบ validation ของ totalIncome และ wht: เคสที่ต้องได้ 400 พร้อมข้อความ error
// และเคสขอบ (รายได้ 0, wht เท่ากับรายได้) ที่ต้องผ่านด้วย 200
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
		{"zero income is allowed", `{"totalIncome":0,"wht":0,"allowances":[]}`, 200, ""},
		{"wht equal to income is allowed", `{"totalIncome":100000,"wht":100000,"allowances":[]}`, 200, ""},
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
			// ทุกเคส 400 ต้องตอบเป็น JSON รูปแบบ {"error":"..."}
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

// ทดสอบว่า response มี taxLevel ครบ 5 ช่วง (รายได้ 750,000)
// แปลง JSON กลับเป็น struct แล้วเทียบด้วย tolerance เพราะเลข float อาจมีเศษทศนิยม
func TestCalculateReturnsTaxLevel(t *testing.T) {
	router := NewRouter()

	body := `{"totalIncome":750000,"wht":0,"allowances":[]}`
	req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got struct {
		Tax      float64 `json:"tax"`
		TaxLevel []struct {
			Level string  `json:"level"`
			Tax   float64 `json:"tax"`
		} `json:"taxLevel"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v (body = %q)", err, rec.Body.String())
	}

	wantLabels := []string{"0-150,000", "150,001-500,000", "500,001-1,000,000", "1,000,001-2,000,000", "2,000,001 ขึ้นไป"}
	wantTaxes := []float64{0, 35000, 28500, 0, 0}

	if math.Abs(got.Tax-63500) > 0.01 {
		t.Errorf("tax = %v, want 63500", got.Tax)
	}
	if len(got.TaxLevel) != len(wantLabels) {
		t.Fatalf("len(taxLevel) = %d, want %d", len(got.TaxLevel), len(wantLabels))
	}
	for i := range wantLabels {
		if got.TaxLevel[i].Level != wantLabels[i] || math.Abs(got.TaxLevel[i].Tax-wantTaxes[i]) > 0.01 {
			t.Errorf("taxLevel[%d] = %+v, want {%s %v}", i, got.TaxLevel[i], wantLabels[i], wantTaxes[i])
		}
	}
}

// ทดสอบการคำนวณเมื่อมีค่าลดหย่อนเงินบริจาค: เกินเพดาน 100,000, มี WHT และ allowances ว่าง
func TestCalculateWithAllowances(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name    string
		body    string
		wantTax float64
	}{
		{"donation over cap", `{"totalIncome":700000,"wht":0,"allowances":[{"allowanceType":"donation","amount":120000}]}`, 41000},
		{"example 4 income, capped", `{"totalIncome":850000,"wht":0,"allowances":[{"allowanceType":"donation","amount":150000}]}`, 63500},
		{"donation with wht", `{"totalIncome":700000,"wht":10000,"allowances":[{"allowanceType":"donation","amount":120000}]}`, 31000},
		{"empty allowances", `{"totalIncome":750000,"wht":0,"allowances":[]}`, 63500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body = %s)", rec.Code, rec.Body.String())
			}
			var got struct {
				Tax float64 `json:"tax"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if math.Abs(got.Tax-tc.wantTax) > 0.01 {
				t.Errorf("tax = %v, want %v", got.Tax, tc.wantTax)
			}
		})
	}
}
