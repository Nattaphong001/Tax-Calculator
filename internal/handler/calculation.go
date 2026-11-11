package handler

import (
	"encoding/json"
	"net/http"

	"tax-calculator/internal/tax"
)

// allowance คือค่าลดหย่อนหนึ่งรายการ
type allowance struct {
	AllowanceType string  `json:"allowanceType"` // ประเภทของค่าลดหย่อน
	Amount        float64 `json:"amount"`        // จำนวนเงินค่าลดหย่อน
}

// calculationRequest คือ JSON ที่ client ส่งเข้ามา
type calculationRequest struct {
	TotalIncome *float64    `json:"totalIncome"` // รายได้รวมทั้งหมด
	WHT         *float64    `json:"wht"`         // ภาษีหัก ณ ที่จ่าย
	Allowances  []allowance `json:"allowances"`  // รายการค่าลดหย่อนภาษี
}

// calculationResponse คือ JSON ที่ Server ตอบกลับไป
type calculationResponse struct {
	Tax float64 `json:"tax"` // ภาษีที่ต้องจ่าย
}

// CalculateHandler รับ POST /tax/calculations
func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req calculationRequest
	// อ่าน JSON จาก request body เข้ามาเก็บใน req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// ส่ง req ไปให้ฟังก์ชัน validate ตรวจค่าตามกฎของโจทย์ก่อนคำนวณ
	if msg := validate(req); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	// ผ่าน validate แล้ว: TotalIncome ไม่เป็น nil แน่นอน อ่านผ่าน * ได้
	// wht (ภาษีหัก ณ ที่จ่าย) เป็น optional ถ้าไม่ส่งมาถือเป็น 0
	wht := 0.0
	if req.WHT != nil {
		wht = *req.WHT
	}

	// เรียก tax.Calculate() เพื่อคำนวณภาษี
	result := tax.Calculate(*req.TotalIncome, wht)

	// ส่ง response กลับไปยัง client เป็น JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calculationResponse{Tax: result})
}
