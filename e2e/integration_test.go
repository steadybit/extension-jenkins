// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extlogging.InitZeroLog()
	server := createMockJenkinsServer()
	defer server.Close()
	split := strings.SplitAfter(server.URL, ":")
	port := split[len(split)-1]

	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-jenkins",
		Port: 8082,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{
				"--set", "logging.level=debug",
				"--set", fmt.Sprintf("jenkins.baseUrl=http://host.minikube.internal:%s", port),
				"--set", "jenkins.apiToken=123xyz",
				"--set", "jenkins.apiUser=admin",
			}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "validate discovery",
			Test: validateDiscovery,
		},
		{
			Name: "target discovery",
			Test: testDiscovery,
		},
	})
}

func validateDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, validate.ValidateEndpointReferences("/", e.Client))
}

func testDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, "com.steadybit.extension_jenkins.job", func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "steadybit.label", "another project")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, "com.steadybit.extension_jenkins.job")
	assert.Equal(t, target.Attributes["jenkins.job.name"], []string{"another project"})
	assert.Equal(t, target.Attributes["jenkins.job.name.full"], []string{"another project"})
	assert.Contains(t, target.Attributes["jenkins.job.parameter"], "Are you sure?")
	assert.Contains(t, target.Attributes["jenkins.job.parameter"], "Say something")

	target, err = e2e.PollForTarget(ctx, e, "com.steadybit.extension_jenkins.job", func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "steadybit.label", "This is a folder » Folder-project")
	})
	require.NoError(t, err)
	assert.Equal(t, target.TargetType, "com.steadybit.extension_jenkins.job")
	assert.Equal(t, target.Attributes["jenkins.job.name"], []string{"Folder-project"})
	assert.Equal(t, target.Attributes["jenkins.job.name.full"], []string{"Folder/Folder-project"})

}
