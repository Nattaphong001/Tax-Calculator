package handler

import (
	"encoding/json"
	"net/http"

	"tax-calculator/internal/tax"
)

// allowance คือค่าลดหย่อนหนึ่งรายการ
type allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

// calculationRequest คือ JSON ที่ client ส่งเข้ามา
type calculationRequest struct {
	TotalIncome float64     `json:"totalIncome"`
	WHT         float64     `json:"wht"`
	Allowances  []allowance `json:"allowances"`
}

// calculationResponse คือ JSON ที่เราตอบกลับ
type calculationResponse struct {
	Tax float64 `json:"tax"`
}

// CalculateHandler รับ POST /tax/calculations
func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req calculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// ชั่วคราว: Step 5 จะทำ error message ให้ชัดเจน
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := tax.Calculate(req.TotalIncome, req.WHT)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calculationResponse{Tax: result})
}