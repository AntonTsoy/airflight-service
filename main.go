// @title My API
// @version 1.0
// @description This is a sample REST API in Go.
// @host localhost:8080
// @BasePath /
package main

import (
	"net/http"

	_ "github.com/AntonTsoy/airflight-service/docs" // Импортируем сгенерированную документацию
	httpSwagger "github.com/swaggo/http-swagger"    // Добавляем http-swagger
)

// @Summary Get a greeting
// @Description Returns a simple greeting
// @Tags example
// @Produce plain
// @Success 200 {string} string "Hello, World!"
// @Router /hello [get]
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Guys!"))
}

func main() {
	// Обычные роуты
	http.HandleFunc("/hello", HelloHandler)

	// Добавляем Swagger UI по адресу /swagger/
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	http.ListenAndServe(":8080", nil)
}
