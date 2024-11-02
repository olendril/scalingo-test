package main

import (
	"fmt"
	"github.com/olendril/scalingo-test/api"
	github_client "github.com/olendril/scalingo-test/internal/github-client"
	"net/http"
	"os"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
)

func main() {
	log := logger.Default()
	log.Info("Initializing app")
	cfg, err := newConfig()
	if err != nil {
		log.WithError(err).Error("Fail to initialize configuration")
		os.Exit(1)
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)

	server := github_client.NewServer(log, cfg.GithubAccessToken)
	if server == nil {
		log.Error("Fail to create github handler")
		os.Exit(2)
	}

	handler := api.HandlerFromMux(*server, router.Router)
	// Initialize web server and configure the following routes:
	// GET /repos
	// GET /stats

	log = log.WithField("port", cfg.Port)
	log.Info("Listening...")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler)
	if err != nil {
		log.WithError(err).Error("Fail to listen to the given port")
		os.Exit(2)
	}
}
