package main

import (
	"flag"

	"github.com/AKovalevich/simplecrawler/internal/app"
	"github.com/AKovalevich/simplecrawler/internal/logger"
)

func main() {
	serverPort := flag.String("port", app.DefaultServerPort, "HTTP server port")
	logLevel := flag.String("log_level", app.DefaultLogLevel, "Log level")
	flag.Parse()

	appLogger, err := logger.New(*logLevel)
	if err != nil {
		panic(err)
	}

	crawlerApp := app.New(appLogger, app.Config{Port: *serverPort})
	crawlerApp.Run()
}
