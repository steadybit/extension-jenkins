package e2e

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

func createMockJenkinsServer() *httptest.Server {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen: %v", err))
	}
	server := httptest.Server{
		Listener: listener,
		Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info().Str("path", r.URL.Path).Str("method", r.Method).Str("query", r.URL.RawQuery).Msg("Request received")

			if strings.HasSuffix(r.URL.Path, "/job/project 1/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getProject1())
			} else if strings.HasSuffix(r.URL.Path, "/job/Folder/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getFolder())
			} else if strings.HasSuffix(r.URL.Path, "/job/Folder/job/Folder-project/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getProjectInFolder())
			} else if strings.HasSuffix(r.URL.Path, "/api/json") && !strings.Contains(r.URL.Path, "/job") {
				w.WriteHeader(http.StatusOK)
				w.Write(getRoot())
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		})},
	}
	server.Start()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server")
	return &server
}

func getFolder() []byte {
	log.Info().Msg("Return folder response")
	return []byte(`{
  "_class": "com.cloudbees.hudson.plugins.folder.Folder",
  "actions": [
    {},
    {},
    {
      "_class": "com.cloudbees.plugins.credentials.ViewCredentialsAction"
    }
  ],
  "description": "",
  "displayName": "This is a folder",
  "displayNameOrNull": "This is a folder",
  "fullDisplayName": "This is a folder",
  "fullName": "Folder",
  "name": "Folder",
  "url": "http://jenkins:8080/job/Folder/",
  "healthReport": [],
  "jobs": [
    {
      "_class": "hudson.model.FreeStyleProject",
      "name": "Folder-project",
      "url": "http://jenkins:8080/job/Folder/job/Folder-project/",
      "color": "notbuilt"
    }
  ],
  "primaryView": {
    "_class": "hudson.model.AllView",
    "name": "All",
    "url": "http://jenkins:8080/job/Folder/"
  },
  "views": [
    {
      "_class": "hudson.model.AllView",
      "name": "All",
      "url": "http://jenkins:8080/job/Folder/"
    }
  ]
}`)
}

func getProjectInFolder() []byte {
	log.Info().Msg("Return project in folder response")
	return []byte(`{
  "_class": "hudson.model.FreeStyleProject",
  "actions": [
    {},
    {},
    {
      "_class": "org.jenkinsci.plugins.displayurlapi.actions.JobDisplayAction"
    },
    {
      "_class": "com.cloudbees.plugins.credentials.ViewCredentialsAction"
    }
  ],
  "description": "Test",
  "displayName": "Folder-project",
  "displayNameOrNull": null,
  "fullDisplayName": "This is a folder » Folder-project",
  "fullName": "Folder/Folder-project",
  "name": "Folder-project",
  "url": "http://jenkins:8080/job/Folder/job/Folder-project/",
  "buildable": true,
  "builds": [],
  "color": "notbuilt",
  "firstBuild": null,
  "healthReport": [],
  "inQueue": false,
  "keepDependencies": false,
  "lastBuild": null,
  "lastCompletedBuild": null,
  "lastFailedBuild": null,
  "lastStableBuild": null,
  "lastSuccessfulBuild": null,
  "lastUnstableBuild": null,
  "lastUnsuccessfulBuild": null,
  "nextBuildNumber": 1,
  "property": [],
  "queueItem": null,
  "concurrentBuild": false,
  "disabled": false,
  "downstreamProjects": [],
  "labelExpression": null,
  "scm": {
    "_class": "hudson.scm.NullSCM"
  },
  "upstreamProjects": []
}`)
}

func getProject1() []byte {
	log.Info().Msg("Return project 1 response")
	return []byte(`{
  "_class": "hudson.model.FreeStyleProject",
  "actions": [
    {
      "_class": "hudson.model.ParametersDefinitionProperty",
      "parameterDefinitions": [
        {
          "_class": "hudson.model.BooleanParameterDefinition",
          "defaultParameterValue": {
            "_class": "hudson.model.BooleanParameterValue",
            "value": false
          },
          "description": null,
          "name": "Are you sure?",
          "type": "BooleanParameterDefinition"
        },
        {
          "_class": "hudson.model.StringParameterDefinition",
          "defaultParameterValue": {
            "_class": "hudson.model.StringParameterValue",
            "value": "example"
          },
          "description": null,
          "name": "Say something",
          "type": "StringParameterDefinition"
        }
      ]
    },
    {},
    {},
    {
      "_class": "org.jenkinsci.plugins.displayurlapi.actions.JobDisplayAction"
    },
    {
      "_class": "com.cloudbees.plugins.credentials.ViewCredentialsAction"
    }
  ],
  "description": "saS",
  "displayName": "another project",
  "displayNameOrNull": null,
  "fullDisplayName": "another project",
  "fullName": "another project",
  "name": "another project",
  "url": "http://jenkins:8080/job/project%201/",
  "buildable": true,
  "builds": [],
  "color": "red",
  "firstBuild": null,
  "healthReport": [],
  "inQueue": false,
  "keepDependencies": false,
  "lastBuild":null,
  "lastCompletedBuild": null,
  "lastFailedBuild": null,
  "lastStableBuild": null,
  "lastSuccessfulBuild": null,
  "lastUnstableBuild": null,
  "lastUnsuccessfulBuild": null,
  "nextBuildNumber": 1,
  "property": [
    {
      "_class": "hudson.model.ParametersDefinitionProperty",
      "parameterDefinitions": [
        {
          "_class": "hudson.model.BooleanParameterDefinition",
          "defaultParameterValue": {
            "_class": "hudson.model.BooleanParameterValue",
            "name": "Are you sure?",
            "value": false
          },
          "description": null,
          "name": "Are you sure?",
          "type": "BooleanParameterDefinition"
        },
        {
          "_class": "hudson.model.StringParameterDefinition",
          "defaultParameterValue": {
            "_class": "hudson.model.StringParameterValue",
            "name": "Say something",
            "value": "beeeeeeeep"
          },
          "description": null,
          "name": "Say something",
          "type": "StringParameterDefinition"
        }
      ]
    }
  ],
  "queueItem": null,
  "concurrentBuild": false,
  "disabled": false,
  "downstreamProjects": [],
  "labelExpression": null,
  "scm": {
    "_class": "hudson.scm.NullSCM"
  },
  "upstreamProjects": []
}`)
}

func getRoot() []byte {
	log.Info().Msg("Return root response")
	return []byte(`{
  "_class": "hudson.model.Hudson",
  "assignedLabels": [
    {
      "name": "built-in"
    }
  ],
  "mode": "NORMAL",
  "nodeDescription": "the Jenkins controller's built-in node",
  "nodeName": "",
  "numExecutors": 0,
  "description": null,
  "jobs": [
    {
      "_class": "hudson.model.FreeStyleProject",
      "name": "project 1",
      "url": "http://jenkins:8080/job/project%201/",
      "color": "red"
    },
    {
      "_class": "com.cloudbees.hudson.plugins.folder.Folder",
      "name": "Folder",
      "url": "http://jenkins:8080/job/Folder/"
    }
  ],
  "overallLoad": {},
  "primaryView": {
    "_class": "hudson.model.AllView",
    "name": "all",
    "url": "http://jenkins:8080/"
  },
  "quietDownReason": null,
  "quietingDown": false,
  "slaveAgentPort": 50000,
  "unlabeledLoad": {
    "_class": "jenkins.model.UnlabeledLoadStatistics"
  },
  "url": "http://jenkins:8080/",
  "useCrumbs": true,
  "useSecurity": true,
  "views": [
    {
      "_class": "hudson.model.AllView",
      "name": "all",
      "url": "http://jenkins:8080/"
    }
  ]
}`)
}
