package handler

import "net/http"

// NewRouter ลงทะเบียน route ทั้งหมดแล้วคืน http.Handler
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HealthHandler)
	mux.HandleFunc("POST /tax/calculations", CalculateHandler)
	return mux
}