package main

import (
	"context"
	"fmt"
	"os"
	"time"

	gindump "github.com/tpkeeper/gin-dump"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.intuit.com/data-curation/go/intuit"
	"github.intuit.com/dev-build/ibp-genai-service/api"
	"github.intuit.com/dev-build/ibp-genai-service/api/v1"
	"github.intuit.com/dev-build/ibp-genai-service/api/v2"
	"github.intuit.com/dev-build/ibp-genai-service/cmd/ibp-genai-service/config"
	"github.intuit.com/dev-build/ibp-genai-service/internal/genai"
)

// Set at build time using linker flags
var Version, SHA string

func init() {
	// Enforce TLS 1.2
	if err := intuit.EnforceTLS12Egress(); err != nil {
		panic(fmt.Errorf("could not enforce TLS 1.2: %w", err))
	}

	// Set up logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339, NoColor: true})
}

func main() {
	log.Info().Msgf("genai-service version: %s, sha: %s", Version, SHA)

	// Load configuration
	cfg, err := config.NewConfigFromPrefix("APP")
	if err != nil {
		log.Fatal().Err(err).Msg("could not load configuration")
	}

	// Instantiate GenAI client
	genAIClient, err := genai.NewClient(
		cfg.ExpressEndpoint,
		cfg.Identity.Endpoint,
		cfg.Identity.AppID,
		cfg.Identity.JobID,
		cfg.GetIdentityAppSecret,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("could not instantiate GenAI client")
	}

	// Splunk client
	splunkClient, err := cfg.NewSplunkClient()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	// Failure analyzer
	failureAnalyzer := genai.NewFailureAnalyzer(genAIClient)

	// User feedback
	userFeedback := genai.NewUserFeedback(splunkClient)

	// Activity
	activity := genai.NewUserActivity(splunkClient)

	// API services
	v1AIService := v1.NewGenAIService(failureAnalyzer, userFeedback)
	v2AIService := v2.NewGenAIService(failureAnalyzer, userFeedback, activity)
	healthService := api.NewHealthService()

	// Set Gin mode to release if not dev
	if cfg.Env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Instantiate router
	router := gin.New()

	// Instantiate Intuit middlewares
	middlewares, err := cfg.IntuitMiddleware(nil, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("could not get default middleware")
	}

	// Register Intuit middlewares
	router.Use(middlewares...)
	router.NoRoute(func(c *gin.Context) { c.Status(404) })

	// Register health handler
	healthService.RegisterHandlers(router)

	// Create v1 group
	v1 := router.Group("/v1")

	// Middleware to dump request/response
	v1.Use(gindump.DumpWithOptions(true, true, true, true, true, func(dumpStr string) {
		log.Info().Msg(dumpStr)
	}))

	// Register handlers
	v1AIService.RegisterHandlers(v1)

	// Create v1 group
	v2 := router.Group("/v2")

	// Middleware to dump request/response
	v2.Use(gindump.DumpWithOptions(true, true, true, true, true, func(dumpStr string) {
		log.Info().Msg(dumpStr)
	}))

	// Register handlers
	v2AIService.RegisterHandlers(v2)

	// Start server
	switch cfg.Env {
	case "dev":
		err = cfg.ListenAndServeLocal(context.Background(), router)
	default:
		// certs are generated during startup in entrypoint.sh
		// Mesh support is available through MESH_ENABLED and MESH_TRAFFIC_PORT environment variables
		err = cfg.ListenAndServeMsaas(context.Background(), router, "./ssl/certificate.crt", "./ssl/certificate.key")
	}

	// Handle errors
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
