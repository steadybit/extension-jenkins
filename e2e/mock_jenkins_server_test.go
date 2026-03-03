package e2e

import (
	"crypto/tls"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

func createMockJenkinsServer() *httptest.Server {
	// Generate self-signed certificate for TLS
	cert, err := generateSelfSignedCert()
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to generate self-signed certificate: %v", err))
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen: %v", err))
	}

	server := httptest.Server{
		Listener: listener,
		Config: &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Info().Str("path", r.URL.Path).Str("method", r.Method).Str("query", r.URL.RawQuery).Msg("Request received")
				baseURL := fmt.Sprintf("https://%s", r.Host)

				if strings.HasSuffix(r.URL.Path, "/job/my-job/api/json") {
					w.WriteHeader(http.StatusOK)
					w.Write(getMyJob(baseURL))
				} else if strings.HasSuffix(r.URL.Path, "/job/Folder/api/json") {
					w.WriteHeader(http.StatusOK)
					w.Write(getFolder(baseURL))
				} else if strings.HasSuffix(r.URL.Path, "/crumbIssuer/api/json/api/json") {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"_class": "hudson.security.csrf.DefaultCrumbIssuer","crumb": "13749c63e9ed3f7dae947786bb7922dcb9f8609a4aba48089cde33b623ab1dc1","crumbRequestField": "Jenkins-Crumb"}`))
				} else if strings.HasSuffix(r.URL.Path, "/job/my-job/buildWithParameters") {
					log.Info().Msg("Return buildWithParameters with Location header")
					w.Header().Add("Location", baseURL+"/queue/item/20/")
					w.WriteHeader(http.StatusOK)
				} else if strings.HasSuffix(r.URL.Path, "/job/Folder/job/Folder-project/api/json") {
					w.WriteHeader(http.StatusOK)
					w.Write(getJobInFolder(baseURL))
				} else if strings.HasSuffix(r.URL.Path, "/queue/item/20/api/json") {
					w.WriteHeader(http.StatusOK)
					w.Write(getQueueItem(baseURL))
				} else if strings.HasSuffix(r.URL.Path, "/api/json") && !strings.Contains(r.URL.Path, "/job") {
					w.WriteHeader(http.StatusOK)
					w.Write(getRoot(baseURL))
				} else if strings.HasSuffix(r.URL.Path, "/job/my-job//9/api/json") {
					w.WriteHeader(http.StatusOK)
					w.Write(getBuild(baseURL))
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			}),
			TLSConfig: &tls.Config{},
		},
		TLS: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	server.StartTLS()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server with self-signed certificate")
	return &server
}

func getBuild(baseURL string) []byte {
	log.Info().Msg("Return build response")
	return []byte(fmt.Sprintf(`{
  "_class": "hudson.model.FreeStyleBuild",
  "actions": [
    {
      "_class": "hudson.model.ParametersAction",
      "parameters": [
        {
          "_class": "hudson.model.BooleanParameterValue",
          "name": "Are you sure?",
          "value": false
        },
        {
          "_class": "hudson.model.StringParameterValue",
          "name": "Just a string input",
          "value": ""
        }
      ]
    },
    {
      "_class": "hudson.model.CauseAction",
      "causes": [
        {
          "_class": "hudson.model.Cause$UserIdCause",
          "shortDescription": "Started by user Jenkins Admin",
          "userId": "admin",
          "userName": "Jenkins Admin"
        }
      ]
    },
    {
      "_class": "jenkins.metrics.impl.TimeInQueueAction",
      "blockedDurationMillis": 0,
      "blockedTimeMillis": 0,
      "buildableDurationMillis": 22938,
      "buildableTimeMillis": 22938,
      "buildingDurationMillis": 15358,
      "executingTimeMillis": 15358,
      "executorUtilization": 1,
      "subTaskCount": 0,
      "waitingDurationMillis": 1,
      "waitingTimeMillis": 1
    },
    {
      "_class": "org.jenkinsci.plugins.displayurlapi.actions.RunDisplayAction"
    }
  ],
  "artifacts": [],
  "building": false,
  "description": null,
  "displayName": "#9",
  "duration": 15358,
  "estimatedDuration": 14848,
  "executor": null,
  "fullDisplayName": "Example Job #9",
  "id": "9",
  "inProgress": false,
  "keepLog": false,
  "number": 9,
  "queueId": 21,
  "result": "SUCCESS",
  "timestamp": 1749733287152,
  "url": "%s/job/my-job/9/",
  "builtOn": "default-zn7ks",
  "changeSet": {
    "_class": "hudson.scm.EmptyChangeLogSet",
    "items": [],
    "kind": null
  },
  "culprits": []
}`, baseURL))
}

func getQueueItem(baseURL string) []byte {
	log.Info().Msg("Return queue item response")
	return []byte(fmt.Sprintf(`{
  "_class": "hudson.model.Queue$LeftItem",
  "actions": [
    {
      "_class": "hudson.model.ParametersAction",
      "parameters": [
        {
          "_class": "hudson.model.BooleanParameterValue",
          "name": "Are you sure?",
          "value": false
        },
        {
          "_class": "hudson.model.StringParameterValue",
          "name": "Just a string input",
          "value": ""
        }
      ]
    },
    {
      "_class": "hudson.model.CauseAction",
      "causes": [
        {
          "_class": "hudson.model.Cause$UserIdCause",
          "shortDescription": "Started by user Jenkins Admin",
          "userId": "admin",
          "userName": "Jenkins Admin"
        }
      ]
    }
  ],
  "blocked": false,
  "buildable": false,
  "id": 21,
  "inQueueSince": 1749733264212,
  "params": "\nAre you sure?=false\nJust a string input=",
  "stuck": false,
  "task": {
    "_class": "hudson.model.FreeStyleProject",
    "name": "my-job",
    "url": "%s/job/my-job/",
    "color": "blue"
  },
  "url": "queue/item/21/",
  "why": null,
  "cancelled": false,
  "executable": {
    "_class": "hudson.model.FreeStyleBuild",
    "number": 9,
    "url": "%s/job/my-job/9/"
  }
}`, baseURL, baseURL))
}
func getFolder(baseURL string) []byte {
	log.Info().Msg("Return folder response")
	return []byte(fmt.Sprintf(`{
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
  "url": "%s/job/Folder/",
  "healthReport": [],
  "jobs": [
    {
      "_class": "hudson.model.FreeStyleProject",
      "name": "Folder-project",
      "url": "%s/job/Folder/job/Folder-project/",
      "color": "notbuilt"
    }
  ],
  "primaryView": {
    "_class": "hudson.model.AllView",
    "name": "All",
    "url": "%s/job/Folder/"
  },
  "views": [
    {
      "_class": "hudson.model.AllView",
      "name": "All",
      "url": "%s/job/Folder/"
    }
  ]
}`, baseURL, baseURL, baseURL, baseURL))
}

func getJobInFolder(baseURL string) []byte {
	log.Info().Msg("Return project in folder response")
	return []byte(fmt.Sprintf(`{
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
  "url": "%s/job/Folder/job/Folder-project/",
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
}`, baseURL))
}

func getMyJob(baseURL string) []byte {
	log.Info().Msg("Return my-job response")
	return []byte(fmt.Sprintf(`{
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
  "displayName": "my-job",
  "displayNameOrNull": null,
  "fullDisplayName": "my-job",
  "fullName": "my-job",
  "name": "my-job",
  "url": "%s/job/my-job/",
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
}`, baseURL))
}

func getRoot(baseURL string) []byte {
	log.Info().Msg("Return root response")
	return []byte(fmt.Sprintf(`{
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
      "name": "my-job",
      "url": "%s/job/my-job/",
      "color": "red"
    },
    {
      "_class": "com.cloudbees.hudson.plugins.folder.Folder",
      "name": "Folder",
      "url": "%s/job/Folder/"
    }
  ],
  "overallLoad": {},
  "primaryView": {
    "_class": "hudson.model.AllView",
    "name": "all",
    "url": "%s/"
  },
  "quietDownReason": null,
  "quietingDown": false,
  "slaveAgentPort": 50000,
  "unlabeledLoad": {
    "_class": "jenkins.model.UnlabeledLoadStatistics"
  },
  "url": "%s/",
  "useCrumbs": true,
  "useSecurity": true,
  "views": [
    {
      "_class": "hudson.model.AllView",
      "name": "all",
      "url": "%s/"
    }
  ]
}`, baseURL, baseURL, baseURL, baseURL, baseURL))
}