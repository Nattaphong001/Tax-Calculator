package handler

import "net/http"

// NewRouter ทำหน้าที่ลงทะเบียนเส้นทาง (Routes) ทั้งหมด และคืนค่ากลับเป็น http.Handler
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HealthHandler)
	mux.HandleFunc("POST /tax/calculations", CalculateHandler)
	return mux
}
