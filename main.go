// @title Airflight Service API
// @version 1.0
// @host localhost:8000
// @BasePath /
package main

import (
	"net/http"

	_ "github.com/AntonTsoy/airflight-service/docs" // Импортируем сгенерированную документацию
	httpSwagger "github.com/swaggo/http-swagger"    // Добавляем http-swagger
)

// @Summary Get a greeting
// @Description Returns a simple greeting
// @Tags exampled
// @Produce plain
// @Success 200 {string} string "Hello, World!"
// @Router /hello [get]
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Guys!"))
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	// Обычные роуты
	http.HandleFunc("/hello", HelloHandler)

	// Добавляем Swagger UI по адресу /swagger/
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	http.ListenAndServe(cfg.ListenAddr, nil)
}
