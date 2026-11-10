package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// w ใช้สำหรับเขียน response กลับไปยัง client
// r ใช้สำหรับอ่านข้อมูล request ที่ client ส่งมา
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") 
	json.NewEncoder(w).Encode(map[string]string{ 
		"status": "ok",
	})
}

func main() {
	// อ่านค่าพอร์ตจาก environment variable "PORT" ถ้าไม่มีให้ใช้ค่า default เป็น "8080"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// สร้าง ServeMux เพื่อจัดการ routing ของ HTTP requests
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}