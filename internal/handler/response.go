package handler

import (
	"encoding/json"
	"net/http"
)

// writeError ตอบ error เป็น JSON รูปแบบเดียวกัน
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
