package main

import (
	"log"
	"net/http"
	"os"

	"tax-calculator/internal/handler"
)

func main() {
	// กำหนด port จาก environment variable หรือใช้ค่า default เป็น 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// log port
	log.Printf("listening on :%s", port)
	// เริ่มต้น HTTP server และใช้ Router จาก handler.NewRouter() ในการจัดเส้นทาง (Route) และจัดการ Request
	log.Fatal(http.ListenAndServe(":"+port, handler.NewRouter()))
}
