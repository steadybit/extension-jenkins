/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"strings"
)

// Specification is the configuration specification for the extension. Configuration values can be applied
// through environment variables. Learn more through the documentation of the envconfig package.
// https://github.com/kelseyhightower/envconfig
type Specification struct {
	// The Jenkins Base Url, like 'https://ci.jenkins.io'
	BaseUrl string `json:"baseUrl" split_words:"true" required:"true"`
	// The Jenkins API User
	ApiUser string `json:"apiUser" split_words:"true" required:"true"`
	// The Jenkins API Token
	ApiToken string `json:"apiToken" split_words:"true" required:"true"`
	// If true, the extension will skip TLS verification when connecting to Jenkins
	InsecureSkipVerify bool `json:"insecureSkipVerify" split_words:"true" required:"false" default:"false"`
	// Timeout for a job to start, otherwise an error is returned
	JobStartTimeoutSeconds int `json:"jobStartTimeoutSeconds" split_words:"true" required:"false" default:"60"`
	// variable STEADYBIT_EXTENSION_DISCOVERY_ATTRIBUTES_EXCLUDES_JOB="jenkins.job.name.full".
	DiscoveryAttributesExcludesJob []string `json:"discoveryAttributesExcludesJob" split_words:"true" required:"false"`
}

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
}

func ValidateConfiguration() {
	// envconfig's `required:"true"` only checks that the variable is *set*: an empty
	// value satisfies it, so the extension would start with a blank configuration and
	// fail much later against the target system. Reject blank values here instead.
	if strings.TrimSpace(Config.BaseUrl) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_BASE_URL must not be empty.")
	}
	if strings.TrimSpace(Config.ApiUser) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_API_USER must not be empty.")
	}
	if strings.TrimSpace(Config.ApiToken) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_API_TOKEN must not be empty.")
	}
}
