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

			if strings.HasSuffix(r.URL.Path, "/job/my-job/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getMyJob())
			} else if strings.HasSuffix(r.URL.Path, "/job/Folder/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getFolder())
			} else if strings.HasSuffix(r.URL.Path, "/crumbIssuer/api/json/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"_class": "hudson.security.csrf.DefaultCrumbIssuer","crumb": "13749c63e9ed3f7dae947786bb7922dcb9f8609a4aba48089cde33b623ab1dc1","crumbRequestField": "Jenkins-Crumb"}`))
			} else if strings.HasSuffix(r.URL.Path, "/job/my-job/buildWithParameters") {
				log.Info().Msg("Return buildWithParameters with Location header")
				w.Header().Add("Location", "http://localhost:8090/queue/item/20/")
				w.WriteHeader(http.StatusOK)
			} else if strings.HasSuffix(r.URL.Path, "/job/Folder/job/Folder-project/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getJobInFolder())
			} else if strings.HasSuffix(r.URL.Path, "/queue/item/20/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getQueueItem())
			} else if strings.HasSuffix(r.URL.Path, "/api/json") && !strings.Contains(r.URL.Path, "/job") {
				w.WriteHeader(http.StatusOK)
				w.Write(getRoot())
			} else if strings.HasSuffix(r.URL.Path, "/job/my-job//9/api/json") {
				w.WriteHeader(http.StatusOK)
				w.Write(getBuild())
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		})},
	}
	server.Start()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server")
	return &server
}

func getBuild() []byte {
	log.Info().Msg("Return build response")
	return []byte(`{
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
  "url": "http://localhost:8090/job/Example%20Job/9/",
  "builtOn": "default-zn7ks",
  "changeSet": {
    "_class": "hudson.scm.EmptyChangeLogSet",
    "items": [],
    "kind": null
  },
  "culprits": []
}`)
}

func getQueueItem() []byte {
	log.Info().Msg("Return queue item response")
	return []byte(`{
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
    "name": "Example Job",
    "url": "http://localhost:8090/job/Example%20Job/",
    "color": "blue"
  },
  "url": "queue/item/21/",
  "why": null,
  "cancelled": false,
  "executable": {
    "_class": "hudson.model.FreeStyleBuild",
    "number": 9,
    "url": "http://localhost:8090/job/Example%20Job/9/"
  }
}`)

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

func getJobInFolder() []byte {
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

func getMyJob() []byte {
	log.Info().Msg("Return my-job response")
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
  "displayName": "my-job",
  "displayNameOrNull": null,
  "fullDisplayName": "my-job",
  "fullName": "my-job",
  "name": "my-job",
  "url": "http://jenkins:8080/job/my-job/",
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
      "name": "my-job",
      "url": "http://jenkins:8080/job/my-job/",
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
