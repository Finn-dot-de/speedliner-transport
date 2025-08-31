// @title          Speedliner API
// @version        1.0
// @description    REST API für Speedliner
// @host           localhost:8080
// @BasePath       /
// @schemes        http
package main

import (
	"log"
	"net/http"
	"os"
	_ "speedliner-server/docs"
	"speedliner-server/src/db"
	"speedliner-server/src/middleware"
	"speedliner-server/src/router"
	"speedliner-server/src/utils"

	httpSwagger "github.com/swaggo/http-swagger"
)

const DefaultAppPort = "8080"

func main() {
	utils.LoadEnv()
	initializeLoggerOrExit("app.log")

	r := router.NewRouter()

	if err := db.InitDB(); err != nil {
		log.Fatal(err)
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = DefaultAppPort
	}

	if os.Getenv("APP_ENV") != "production" {
		r.Handle("/swagger/*", httpSwagger.WrapHandler)
	}

	handler := middleware.LoggerMiddleware(middleware.NoCacheMiddleware(middleware.RateLimit(r)))

	log.Println("🚀 Server läuft auf Port " + appPort)
	log.Println(":" + appPort)
	log.Fatal(http.ListenAndServe(":"+appPort, handler))

}

func initializeLoggerOrExit(logFile string) {
	err := middleware.InitializeLogger(logFile)
	if err != nil {
		log.Fatalf("Fehler beim Initialisieren des Loggings: %v", err)
	}
}
