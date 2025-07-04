/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package main

import (
	"context"
	"crypto/tls"
	_ "github.com/KimMachineGun/automemlimit" // By default, it sets `GOMEMLIMIT` to 90% of cgroup's memory limit.
	"github.com/bndr/gojenkins"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-jenkins/config"
	"github.com/steadybit/extension-jenkins/extjenkins"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-kit/extsignals"
	_ "go.uber.org/automaxprocs" // Importing automaxprocs automatically adjusts GOMAXPROCS.
	"net/http"
)

func main() {
	extlogging.InitZeroLog()

	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)

	config.ParseConfiguration()
	config.ValidateConfiguration()

	exthealth.SetReady(false)
	exthealth.StartProbes(8083)

	ctx := context.Background()

	// Create HTTP client with optional TLS configuration for self-signed certificates
	var httpClient *http.Client
	if config.Config.InsecureSkipVerify {
		log.Info().Msg("TLS verification disabled for Jenkins connection. Self-signed certificates will be accepted.")
		insecureTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //NOSONAR explicit choice
		}
		httpClient = &http.Client{Transport: insecureTransport}
	}

	jenkins := gojenkins.CreateJenkins(httpClient, config.Config.BaseUrl, config.Config.ApiUser, config.Config.ApiToken)
	_, err := jenkins.Init(ctx)

	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to connect to Jenkins at %s", config.Config.BaseUrl)
	}

	discovery_kit_sdk.Register(extjenkins.NewJobDiscovery(jenkins))
	action_kit_sdk.RegisterAction(extjenkins.NewJobRunAction(jenkins))

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))

	extsignals.ActivateSignalHandlers()

	action_kit_sdk.RegisterCoverageEndpoints()
	exthealth.SetReady(true)

	exthttp.Listen(exthttp.ListenOpts{
		Port: 8082,
	})
}

type ExtensionListResponse struct {
	action_kit_api.ActionList       `json:",inline"`
	discovery_kit_api.DiscoveryList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList:    action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_sdk.GetDiscoveryList(),
	}
}
