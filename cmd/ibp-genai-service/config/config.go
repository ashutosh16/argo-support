package config

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.intuit.com/data-curation/go/config"
	"github.intuit.com/dev-build/ibp-genai-service/internal/splunk"
)

const (
	AssetAlias         = "Intuit.dev.build.ibpgenaiservice"
	PrivateAuthIDPSKey = "ibp-genai-service-private-auth" // the secret comes from https://devportal.intuit.com/app/dp/resource/1732777549890492173/credentials/secrets
	ExperienceID       = "d5ea1279-a30f-4844-8bce-3dbf8498c48b"
)

type Config struct {
	config.Config

	ExpressEndpoint string `envconfig:"EXPRESS_ENDPOINT"`
	Identity        struct {
		Endpoint  string `envconfig:"ENDPOINT"`
		AppID     string `envconfig:"APP_ID"`
		AppSecret string `envconfig:"APP_SECRET"`
		JobID     string `envconfig:"JOB_ID"`
	}
	Splunk struct {
		Hostname string `envconfig:"HOSTNAME"`
		Token    string `envconfig:"TOKEN"`
	}
	Timeout string
}

func NewConfigFromPrefix(prefix string) (*Config, error) {
	var cfg Config

	// Set the app ID to the asset alias
	cfg.Identity.AppID = AssetAlias

	// Read configuration from the environment (including IDPS secrets)
	if err := readConfiguration(prefix, &cfg); err != nil {
		log.Error().Err(err).Msg("unable to read secrets from IDPS")
		return nil, err
	}

	// Validate the configuration
	if cfg.ExpressEndpoint == "" {
		return nil, fmt.Errorf("express endpoint is required")
	}
	if cfg.Identity.Endpoint == "" {
		return nil, fmt.Errorf("identity endpoint is required")
	}
	if cfg.Identity.JobID == "" {
		return nil, fmt.Errorf("identity job ID is required")
	}
	if cfg.Identity.AppID == "" {
		return nil, fmt.Errorf("identity app ID is required")
	}
	if cfg.Splunk.Hostname == "" {
		return nil, fmt.Errorf("splunk hostname is required")
	}
	if cfg.Splunk.Token == "" {
		return nil, fmt.Errorf("splunk token is required")
	}
	log.Info().Msg("configuration is valid")

	// Print the configuration
	log.Info().Msgf("Express Endpoint: %s", cfg.ExpressEndpoint)
	log.Info().Msgf("Identity Endpoint: %s", cfg.Identity.Endpoint)
	log.Info().Msgf("Identity JobID: %s", cfg.Identity.JobID)
	log.Info().Msgf("Identity AppID: %s", cfg.Identity.AppID)
	log.Info().Msgf("IDPS endpoint: %s", cfg.Idps.Endpoint)
	log.Info().Msgf("IDPS policy: %s", cfg.Idps.Policy)
	log.Info().Msgf("IDPS folder: %s", cfg.Idps.Folder)
	log.Info().Msgf("Splunk Hostname: %s", cfg.Splunk.Hostname)
	log.Info().Msgf("Splunk Token: %s", printSecret(cfg.Splunk.Token))
	log.Info().Msgf("Identity App Secret: %s", printSecret(cfg.Identity.AppSecret))

	// Reload the configuration every 5 minutes
	go func() {
		for range time.Tick(5 * time.Minute) {
			if err := readConfiguration(prefix, &cfg); err != nil {
				log.Error().Err(err).Msg("unable to refresh configuration")
				continue
			}
			log.Info().Msgf("Successfully refreshed Identity App Secret: %s", printSecret(cfg.Identity.AppSecret))
		}
	}()

	return &cfg, nil
}

func readConfiguration(prefix string, cfg *Config) error {
	// IDPS Secrets are not available from the config.Config struct seemingly because of a bug in the curation library
	// The workaround is to read the secrets into a map[string]string and then store them in the config.Config struct
	secrets := map[string]string{}
	if err := config.Load(prefix, AssetAlias, cfg, &secrets); err != nil {
		return fmt.Errorf("unable to load configuration: %w", err)
	}

	// Retrieve the IDPS secrets we need and store then in the configuration
	appSecret, found := secrets[PrivateAuthIDPSKey]
	if !found {
		return fmt.Errorf("unable to find IDPS key %s", PrivateAuthIDPSKey)
	}

	// Decode base 64 encoded secret
	decoded, err := base64.StdEncoding.DecodeString(appSecret)
	if err != nil {
		return fmt.Errorf("unable to decode secret: %s", printSecret(appSecret))
	}

	// Set the secret in the configuration
	cfg.Identity.AppSecret = string(decoded)

	return nil
}

func printSecret(secret string) string {
	if len(secret) <= 4 {
		return strings.Repeat("*", len(secret))
	}
	return strings.Repeat("*", len(secret)-4) + secret[len(secret)-4:]
}

func (cfg *Config) GetIdentityAppSecret() string {
	return cfg.Identity.AppSecret
}

func (cfg *Config) NewSplunkClient() (*splunk.Splunk, error) {
	splunkUrl := fmt.Sprintf("https://%s/services/collector", cfg.Splunk.Hostname)
	return splunk.NewSplunk(splunkUrl, cfg.Splunk.Token, cfg.Env), nil
}
