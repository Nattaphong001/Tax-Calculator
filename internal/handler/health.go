package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler ตรวจสอบและส่งกลับสถานะการทำงานของเซิร์ฟเวอร์
// w เขียน response กลับไปยัง client , r คือ request ที่ client ส่งมา
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// กำหนด Content-Type สำหรับ response เป็น application/json
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
