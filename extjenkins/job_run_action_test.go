/*
 * Copyright 2026 steadybit GmbH. All rights reserved.
 */

package extjenkins

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jenkinsMock is a minimal Jenkins stand-in for the handful of endpoints the run
// action touches through gojenkins: the job, its queue item, the build, and the
// POSTs that trigger, cancel and stop it. State is mutable so a test can walk a
// job from queued to finished.
type jenkinsMock struct {
	server *httptest.Server

	mu sync.Mutex
	// jobParameters are the names of the job's parameter definitions.
	jobParameters []string
	// executable is the build number the queue item resolved to, 0 while queued.
	executable int64
	building   bool
	result     string
	// missingJob makes every job lookup a 404.
	missingJob bool
	// requests records "METHOD /path" of every request seen.
	requests []string
}

func newJenkinsMock(t *testing.T) *jenkinsMock {
	t.Helper()
	m := &jenkinsMock{}
	m.server = httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(m.server.Close)
	return m
}

func (m *jenkinsMock) client() *gojenkins.Jenkins {
	return gojenkins.CreateJenkins(nil, m.server.URL, "user", "token")
}

func (m *jenkinsMock) set(fn func(m *jenkinsMock)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m)
}

func (m *jenkinsMock) saw(prefix string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.requests {
		if strings.HasPrefix(r, prefix) {
			return true
		}
	}
	return false
}

func (m *jenkinsMock) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// gojenkins builds the build path from the job url, which yields a double slash.
	path := strings.ReplaceAll(r.URL.Path, "//", "/")
	m.requests = append(m.requests, r.Method+" "+path)

	writeJson := func(v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}

	switch {
	case path == "/crumbIssuer/api/json/api/json":
		writeJson(map[string]string{"crumbRequestField": "Jenkins-Crumb", "crumb": "crumb"})

	case path == "/job/my-job/api/json":
		if m.missingJob {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		definitions := make([]map[string]any, 0, len(m.jobParameters))
		for _, p := range m.jobParameters {
			definitions = append(definitions, map[string]any{"name": p, "type": "StringParameterDefinition"})
		}
		writeJson(map[string]any{
			"name":     "my-job",
			"url":      m.server.URL + "/job/my-job/",
			"inQueue":  false,
			"property": []map[string]any{{"parameterDefinitions": definitions}},
		})

	case r.Method == http.MethodPost && (path == "/job/my-job/build" || path == "/job/my-job/buildWithParameters"):
		w.Header().Set("Location", m.server.URL+"/queue/item/20/")
		w.WriteHeader(http.StatusCreated)

	case path == "/queue/item/20/api/json":
		writeJson(map[string]any{
			"id":         20,
			"executable": map[string]any{"number": m.executable, "url": m.server.URL + "/job/my-job/9/"},
		})

	case r.Method == http.MethodPost && path == "/queue/cancelItem":
		w.WriteHeader(http.StatusOK)

	case path == "/job/my-job/9/api/json":
		writeJson(map[string]any{
			"number":   9,
			"url":      m.server.URL + "/job/my-job/9/",
			"building": m.building,
			"result":   m.result,
		})

	case r.Method == http.MethodPost && path == "/job/my-job/9/stop":
		m.building = false
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func jobPrepareRequest(cfg map[string]any, jobParameters ...string) action_kit_api.PrepareActionRequestBody {
	attributes := map[string][]string{
		"jenkins.job.name":      {"my-job"},
		"jenkins.job.name.full": {"my-job"},
	}
	if len(jobParameters) > 0 {
		attributes["jenkins.job.parameter"] = jobParameters
	}
	return action_kit_api.PrepareActionRequestBody{
		Config: cfg,
		Target: &action_kit_api.Target{Attributes: attributes},
	}
}

func keyValue(pairs ...string) []any {
	kv := make([]any, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		kv = append(kv, map[string]any{"key": pairs[i], "value": pairs[i+1]})
	}
	return kv
}

func TestExtractParentIds(t *testing.T) {
	assert.Equal(t, []string{}, extractParentIds("my-job"))
	assert.Equal(t, []string{"folder"}, extractParentIds("folder/my-job"))
	assert.Equal(t, []string{"a", "b"}, extractParentIds("a/b/my-job"))
}

func TestJobRunPrepare(t *testing.T) {
	action := &jobRunAction{}

	t.Run("without parameters", func(t *testing.T) {
		state := action.NewEmptyState()
		result, err := action.Prepare(context.Background(), &state, jobPrepareRequest(map[string]any{"waitForCompletion": true}))
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "my-job", state.JobName)
		assert.Equal(t, []string{}, state.ParentIds)
		assert.True(t, state.WaitForCompletion)
		assert.Nil(t, state.Parameters)
	})

	t.Run("with parameters the job defines", func(t *testing.T) {
		state := action.NewEmptyState()
		cfg := map[string]any{"parameters": keyValue("branch", "main")}
		result, err := action.Prepare(context.Background(), &state, jobPrepareRequest(cfg, "branch", "target"))
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, map[string]string{"branch": "main"}, state.Parameters)
	})

	t.Run("warns when the job has no parameters at all", func(t *testing.T) {
		state := action.NewEmptyState()
		cfg := map[string]any{"parameters": keyValue("branch", "main")}
		result, err := action.Prepare(context.Background(), &state, jobPrepareRequest(cfg))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Contains(t, (*result.Messages)[0].Message, "does not have any parameters defined")
	})

	t.Run("warns about parameters the job does not define", func(t *testing.T) {
		state := action.NewEmptyState()
		cfg := map[string]any{"parameters": keyValue("branch", "main", "typo", "x")}
		result, err := action.Prepare(context.Background(), &state, jobPrepareRequest(cfg, "branch"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "- ⚠️ The following parameters are not defined for this job: typo", (*result.Messages)[0].Message)
	})

	t.Run("fails on malformed parameters", func(t *testing.T) {
		state := action.NewEmptyState()
		cfg := map[string]any{"parameters": "not-a-list"}
		_, err := action.Prepare(context.Background(), &state, jobPrepareRequest(cfg, "branch"))
		require.Error(t, err)
	})
}

func TestJobRunStart(t *testing.T) {
	t.Run("queues the job", func(t *testing.T) {
		m := newJenkinsMock(t)
		action := &jobRunAction{jenkins: m.client()}
		state := JobRunActionState{JobName: "my-job", ParentIds: []string{}}

		result, err := action.Start(context.Background(), &state)
		require.NoError(t, err)
		assert.Equal(t, int64(20), state.QueueId)
		assert.Equal(t, "- Waiting for job to start...", (*result.Messages)[0].Message)
		assert.True(t, m.saw("POST /job/my-job/build"))
	})

	t.Run("queues the job with parameters", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.jobParameters = []string{"branch"} })
		action := &jobRunAction{jenkins: m.client()}
		state := JobRunActionState{JobName: "my-job", Parameters: map[string]string{"branch": "main"}}

		_, err := action.Start(context.Background(), &state)
		require.NoError(t, err)
		assert.True(t, m.saw("POST /job/my-job/buildWithParameters"))
	})

	t.Run("fails when the job does not exist", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.missingJob = true })
		action := &jobRunAction{jenkins: m.client()}
		state := JobRunActionState{JobName: "my-job"}

		_, err := action.Start(context.Background(), &state)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to find job.")
	})

	t.Run("fails when jenkins is unreachable while queueing", func(t *testing.T) {
		action := &jobRunAction{jenkins: gojenkins.CreateJenkins(nil, "http://127.0.0.1:1", "user", "token")}
		state := JobRunActionState{JobName: "my-job"}

		_, err := action.Start(context.Background(), &state)
		require.Error(t, err)
	})
}

func TestJobRunStatus(t *testing.T) {
	newAction := func(m *jenkinsMock) *jobRunAction { return &jobRunAction{jenkins: m.client()} }
	farFuture := time.Since(referenceTime) + time.Hour

	t.Run("keeps waiting while the job is queued", func(t *testing.T) {
		m := newJenkinsMock(t)
		state := JobRunActionState{JobName: "my-job", QueueId: 20, TimeoutOffset: farFuture}
		result, err := newAction(m).Status(context.Background(), &state)
		require.NoError(t, err)
		assert.False(t, result.Completed)
		assert.Nil(t, result.Error)
		assert.Zero(t, state.RunId)
	})

	t.Run("times out while the job is queued", func(t *testing.T) {
		m := newJenkinsMock(t)
		state := JobRunActionState{JobName: "my-job", QueueId: 20, TimeoutOffset: 0}
		result, err := newAction(m).Status(context.Background(), &state)
		require.NoError(t, err)
		assert.True(t, result.Completed)
		require.NotNil(t, result.Error)
		assert.Equal(t, "Timed out waiting for job to start.", result.Error.Title)
		assert.Equal(t, action_kit_api.Errored, *result.Error.Status)
	})

	t.Run("completes right away when not waiting for completion", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9; m.building = true })
		state := JobRunActionState{JobName: "my-job", QueueId: 20, WaitForCompletion: false, TimeoutOffset: farFuture}
		result, err := newAction(m).Status(context.Background(), &state)
		require.NoError(t, err)
		assert.True(t, result.Completed)
		assert.True(t, state.DontStop)
		assert.Equal(t, int64(9), state.RunId)
		assert.Contains(t, (*result.Messages)[0].Message, "action will not wait for completion")
	})

	t.Run("reports the start and keeps waiting for completion", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9; m.building = true })
		state := JobRunActionState{JobName: "my-job", QueueId: 20, WaitForCompletion: true, TimeoutOffset: farFuture}
		action := newAction(m)

		result, err := action.Status(context.Background(), &state)
		require.NoError(t, err)
		assert.False(t, result.Completed)
		assert.Equal(t, int64(9), state.RunId)
		assert.Contains(t, (*result.Messages)[0].Message, "- Job started.")

		// Still building: no new message, not completed.
		result, err = action.Status(context.Background(), &state)
		require.NoError(t, err)
		assert.False(t, result.Completed)
		assert.Nil(t, result.Messages)

		// Finished successfully.
		m.set(func(m *jenkinsMock) { m.building = false; m.result = gojenkins.STATUS_SUCCESS })
		result, err = action.Status(context.Background(), &state)
		require.NoError(t, err)
		assert.True(t, result.Completed)
		assert.Nil(t, result.Error)
		assert.True(t, state.DontStop)
		assert.Equal(t, "- Job ended with result 'SUCCESS' ✅", (*result.Messages)[0].Message)
	})

	t.Run("fails the action when the build failed", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9; m.building = false; m.result = "FAILURE" })
		state := JobRunActionState{JobName: "my-job", QueueId: 20, RunId: 9, WaitForCompletion: true, TimeoutOffset: farFuture}
		result, err := newAction(m).Status(context.Background(), &state)
		require.NoError(t, err)
		assert.True(t, result.Completed)
		require.NotNil(t, result.Error)
		assert.Equal(t, action_kit_api.Failed, *result.Error.Status)
		assert.Equal(t, "Job ended with result: FAILURE", result.Error.Title)
		assert.Equal(t, "- Job ended with result 'FAILURE' ⚠️", (*result.Messages)[0].Message)
	})

	t.Run("fails when the queue item cannot be fetched", func(t *testing.T) {
		action := &jobRunAction{jenkins: gojenkins.CreateJenkins(nil, "http://127.0.0.1:1", "user", "token")}
		state := JobRunActionState{JobName: "my-job", QueueId: 20}
		_, err := action.Status(context.Background(), &state)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to fetch task.")
	})

	t.Run("fails when the job vanished", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9; m.missingJob = true })
		state := JobRunActionState{JobName: "my-job", QueueId: 20}
		_, err := newAction(m).Status(context.Background(), &state)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to find job.")
	})

	t.Run("fails when the build cannot be fetched", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9 })
		// RunId 7 has no build behind it in the mock.
		state := JobRunActionState{JobName: "my-job", QueueId: 20, RunId: 7}
		_, err := newAction(m).Status(context.Background(), &state)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to fetch build.")
	})
}

func TestJobRunStop(t *testing.T) {
	newAction := func(m *jenkinsMock) *jobRunAction { return &jobRunAction{jenkins: m.client()} }

	t.Run("does nothing when the job already ended", func(t *testing.T) {
		m := newJenkinsMock(t)
		result, err := newAction(m).Stop(context.Background(), &JobRunActionState{DontStop: true})
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.False(t, m.saw("POST"))
	})

	t.Run("cancels a queued job", func(t *testing.T) {
		m := newJenkinsMock(t)
		state := JobRunActionState{JobName: "my-job", QueueId: 20}
		result, err := newAction(m).Stop(context.Background(), &state)
		require.NoError(t, err)
		assert.Empty(t, *result.Messages)
		assert.True(t, m.saw("POST /queue/cancelItem"))
		assert.False(t, m.saw("POST /job/my-job/9/stop"))
	})

	t.Run("stops a running build", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9; m.building = true })
		state := JobRunActionState{JobName: "my-job", QueueId: 20}
		result, err := newAction(m).Stop(context.Background(), &state)
		require.NoError(t, err)
		assert.Equal(t, "- Job stopped. 🛑", (*result.Messages)[0].Message)
		assert.True(t, m.saw("POST /job/my-job/9/stop"))
	})

	t.Run("fails when the queue item cannot be fetched", func(t *testing.T) {
		action := &jobRunAction{jenkins: gojenkins.CreateJenkins(nil, "http://127.0.0.1:1", "user", "token")}
		_, err := action.Stop(context.Background(), &JobRunActionState{JobName: "my-job", QueueId: 20})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to fetch task.")
	})

	t.Run("fails when the job vanished", func(t *testing.T) {
		m := newJenkinsMock(t)
		m.set(func(m *jenkinsMock) { m.executable = 9; m.missingJob = true })
		_, err := newAction(m).Stop(context.Background(), &JobRunActionState{JobName: "my-job", QueueId: 20})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to find job.")
	})
}

func TestJobRunDescribe(t *testing.T) {
	action := NewJobRunAction(nil)
	description := action.Describe()
	assert.Equal(t, TargetTypeJob+".run", description.Id)
	assert.Equal(t, action_kit_api.TimeControlInternal, description.TimeControl)
	assert.Equal(t, JobRunActionState{}, action.NewEmptyState())
}
