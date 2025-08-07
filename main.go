// @title          Speedliner API
// @version        1.0
// @description    REST API für Speedliner
// @host           localhost:8080
// @BasePath       /
// @schemes        http
package main

import (
	"github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"speedliner-server/db"
	_ "speedliner-server/docs"
	"speedliner-server/src/middleware"
	"speedliner-server/src/router"
	"speedliner-server/src/utils"
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

	log.Println("🚀 Server läuft auf Port " + appPort)
	log.Println("http://0.0.0.0:" + appPort)
	log.Fatal(http.ListenAndServe(":"+appPort, r))
}

func initializeLoggerOrExit(logFile string) {
	err := middleware.InitializeLogger(logFile)
	if err != nil {
		log.Fatalf("Fehler beim Initialisieren des Loggings: %v", err)
	}
}
