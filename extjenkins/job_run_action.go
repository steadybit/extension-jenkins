/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extjenkins

import (
	"context"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-jenkins/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"strings"
	"time"
)

type jobRunAction struct {
	jenkins *gojenkins.Jenkins
}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[JobRunActionState]           = (*jobRunAction)(nil)
	_ action_kit_sdk.ActionWithStatus[JobRunActionState] = (*jobRunAction)(nil)
	_ action_kit_sdk.ActionWithStop[JobRunActionState]   = (*jobRunAction)(nil)
)

type JobRunActionState struct {
	JobName           string
	ParentIds         []string
	WaitForCompletion bool
	Parameters        map[string]string
	QueueId           int64
	RunId             int64
	DontStop          bool
	TimeoutOffset     time.Duration `json:"timeoutOffset"`
}

var referenceTime = time.Now()

func NewJobRunAction(jenkins *gojenkins.Jenkins) action_kit_sdk.Action[JobRunActionState] {
	return &jobRunAction{jenkins: jenkins}
}

func (l *jobRunAction) NewEmptyState() JobRunActionState {
	return JobRunActionState{}
}

func (l *jobRunAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.run", TargetTypeJob),
		Label:       "Run Jenkins Job",
		Description: "Starts a Jenkins job.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(TargetIconJob),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType: TargetTypeJob,
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label: "job name",
					Query: "jenkins.job.name=\"\"",
				},
			}),
			QuantityRestriction: extutil.Ptr(action_kit_api.ExactlyOne),
		}),
		Technology:  extutil.Ptr("Jenkins"),
		Kind:        action_kit_api.Other,
		TimeControl: action_kit_api.TimeControlInternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:         "duration",
				Label:        "Estimated Duration",
				Description:  extutil.Ptr("If `Wait for Completion` is checked, the step will run as long as needed. You can set this estimation to size the step in the experiment editor for a better understanding of the time schedule."),
				Type:         action_kit_api.ActionParameterTypeDuration,
				DefaultValue: extutil.Ptr("60s"),
				Required:     extutil.Ptr(true),
			},
			{
				Name:         "waitForCompletion",
				Label:        "Wait for Completion",
				Description:  extutil.Ptr("If enabled, the action will wait for the job to complete before returning. If disabled, the action will return immediately after starting the job."),
				Type:         action_kit_api.ActionParameterTypeBoolean,
				DefaultValue: extutil.Ptr("true"),
				Required:     extutil.Ptr(true),
			},
			{
				Name:        "parameters",
				Label:       "Parameters",
				Description: extutil.Ptr("Optional parameters to pass to the job."),
				Type:        action_kit_api.ActionParameterTypeKeyValue,
				Required:    extutil.Ptr(false),
			},
		},
		Status: extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("2s"),
		}),
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
		Widgets: extutil.Ptr([]action_kit_api.Widget{
			action_kit_api.MarkdownWidget{
				Type:        action_kit_api.ComSteadybitWidgetMarkdown,
				Title:       "Jenkins",
				MessageType: "JENKINS",
				Append:      true,
			},
		}),
	}
}

func (l *jobRunAction) Prepare(_ context.Context, state *JobRunActionState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	state.JobName = extutil.MustHaveValue(request.Target.Attributes, "jenkins.job.name")[0]
	state.ParentIds = extractParentIds(extutil.MustHaveValue(request.Target.Attributes, "jenkins.job.name.full")[0])
	state.WaitForCompletion = extutil.ToBool(request.Config["waitForCompletion"])
	jobStartTimeout := time.Duration(int(time.Second) * config.Config.JobStartTimeoutSeconds)
	state.TimeoutOffset = time.Since(referenceTime) + jobStartTimeout
	if (request.Config["parameters"]) != nil {
		var err error
		state.Parameters, err = extutil.ToKeyValue(request.Config, "parameters")
		if err != nil {
			return nil, err
		}

		availableParameters, hasParams := request.Target.Attributes["jenkins.job.parameter"]
		if (!hasParams || len(availableParameters) == 0) && len(state.Parameters) > 0 {
			return &action_kit_api.PrepareResult{
				Messages: &[]action_kit_api.Message{
					{
						Message: "- ⚠️ This job does not have any parameters defined, but parameters were provided.",
						Type:    extutil.Ptr("JENKINS"),
					},
				},
			}, nil
		}
		missingParameters := []string{}
		for key := range state.Parameters {
			found := false
			for _, param := range availableParameters {
				if key == param {
					found = true
					break
				}
			}
			if !found {
				missingParameters = append(missingParameters, key)
			}
		}
		if len(missingParameters) > 0 {
			return &action_kit_api.PrepareResult{
				Messages: &[]action_kit_api.Message{
					{
						Message: fmt.Sprintf("- ⚠️ The following parameters are not defined for this job: %s", strings.Join(missingParameters, ", ")),
						Type:    extutil.Ptr("JENKINS"),
					},
				},
			}, nil
		}
	}
	return nil, nil
}

func extractParentIds(fullName string) []string {
	if !strings.Contains(fullName, "/") {
		return []string{}
	}
	parts := strings.Split(fullName, "/")
	if len(parts) <= 1 {
		return []string{}
	}
	return parts[:len(parts)-1]
}

func (l *jobRunAction) Start(ctx context.Context, state *JobRunActionState) (*action_kit_api.StartResult, error) {
	log.Info().Str("jobName", state.JobName).Strs("parentIds", state.ParentIds).Msg("Starting job.")

	job, err := l.jenkins.GetJob(ctx, state.JobName, state.ParentIds...)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find job.", err)
	}

	queueId, err := job.InvokeSimple(ctx, state.Parameters)
	if err != nil {
		return nil, extension_kit.ToError("Failed to queue job.", err)
	}
	log.Info().Int64("queueId", queueId).Msg("Job queued successfully.")
	state.QueueId = queueId

	result := &action_kit_api.StartResult{
		Messages: &[]action_kit_api.Message{
			{
				Message: "- Waiting for job to start...",
				Type:    extutil.Ptr("JENKINS"),
			},
		},
	}
	return result, nil
}

func (l *jobRunAction) Status(ctx context.Context, state *JobRunActionState) (*action_kit_api.StatusResult, error) {
	task, err := l.jenkins.GetQueueItem(ctx, state.QueueId)
	if err != nil {
		return nil, extension_kit.ToError("Failed to fetch task.", err)
	}

	justStarted := false
	if state.RunId == 0 && task.Raw.Executable.Number != 0 {
		state.RunId = task.Raw.Executable.Number
		justStarted = true
	}

	if state.RunId != 0 {
		job, err := l.jenkins.GetJob(ctx, state.JobName, state.ParentIds...)
		if err != nil {
			return nil, extension_kit.ToError("Failed to find job.", err)
		}
		build, err := job.GetBuild(ctx, state.RunId)
		if err != nil {
			return nil, extension_kit.ToError("Failed to fetch build.", err)
		}

		if justStarted {
			if !state.WaitForCompletion {
				log.Info().Int64("runId", task.Raw.Executable.Number).Msg("Job started, action will not waiting for completion.")
				state.DontStop = true
				return &action_kit_api.StatusResult{
					Completed: true,
					Messages: &[]action_kit_api.Message{
						{
							Message: fmt.Sprintf("- Job started, action will not wait for completion. [Build](%s) [Console](%sconsole)", build.Raw.URL, build.Raw.URL),
							Type:    extutil.Ptr("JENKINS"),
						},
					},
				}, nil
			} else {
				log.Info().Int64("runId", task.Raw.Executable.Number).Msg("Job started.")
				return &action_kit_api.StatusResult{
					Completed: false,
					Messages: &[]action_kit_api.Message{
						{
							Message: fmt.Sprintf("- Job started. [Open](%s) [Console](%sconsole)", build.Raw.URL, build.Raw.URL),
							Type:    extutil.Ptr("JENKINS"),
						},
					},
				}, nil
			}
		}

		if !build.Raw.Building {
			log.Info().Str("result", build.Raw.Result).Msg("Job completed.")
			state.DontStop = true
			var result *action_kit_api.ActionKitError = nil
			var messages []action_kit_api.Message
			if build.Raw.Result != gojenkins.STATUS_FIXED && build.Raw.Result != gojenkins.STATUS_SUCCESS && build.Raw.Result != gojenkins.STATUS_PASSED {
				result = &action_kit_api.ActionKitError{
					Status: extutil.Ptr(action_kit_api.Failed),
					Title:  fmt.Sprintf("Job ended with result: %s", build.Raw.Result),
				}
				messages = append(messages, action_kit_api.Message{
					Message: fmt.Sprintf("- Job ended with result '%s' ⚠️", build.Raw.Result),
					Type:    extutil.Ptr("JENKINS"),
				})
			} else {
				messages = append(messages, action_kit_api.Message{
					Message: fmt.Sprintf("- Job ended with result '%s' ✅", build.Raw.Result),
					Type:    extutil.Ptr("JENKINS"),
				})
			}
			return &action_kit_api.StatusResult{
				Completed: true,
				Error:     result,
				Messages:  &messages,
			}, nil
		}
	} else {
		if time.Since(referenceTime) > state.TimeoutOffset {
			return extutil.Ptr(action_kit_api.StatusResult{
				Completed: true,
				Error: extutil.Ptr(action_kit_api.ActionKitError{
					Title:  "Timed out waiting for job to start.",
					Status: extutil.Ptr(action_kit_api.Errored),
				}),
			}), nil
		}
		log.Info().Msg("Job is queued, waiting for start.")
	}

	return &action_kit_api.StatusResult{
		Completed: false,
	}, nil
}

func (l *jobRunAction) Stop(ctx context.Context, state *JobRunActionState) (*action_kit_api.StopResult, error) {
	if state.DontStop {
		return nil, nil
	}

	task, err := l.jenkins.GetQueueItem(ctx, state.QueueId)
	if err != nil {
		return nil, extension_kit.ToError("Failed to fetch task.", err)
	}
	canceled, err := task.Cancel(ctx)
	if err != nil {
		return nil, extension_kit.ToError("Failed to cancel the task.", err)
	}
	if canceled {
		log.Info().Msg("Task canceled.")
	}

	var messages []action_kit_api.Message
	if task.Raw.Executable.Number != 0 {
		job, err := l.jenkins.GetJob(ctx, state.JobName, state.ParentIds...)
		if err != nil {
			return nil, extension_kit.ToError("Failed to find job.", err)
		}
		build, err := job.GetBuild(ctx, task.Raw.Executable.Number)
		if err != nil {
			return nil, extension_kit.ToError("Failed to fetch build.", err)
		}
		stopped, err := build.Stop(ctx)
		if err != nil {
			return nil, extension_kit.ToError("Failed to stop build.", err)
		}
		if stopped {
			log.Info().Msg("Job stopped.")
			messages = append(messages, action_kit_api.Message{
				Message: "- Job stopped. 🛑",
				Type:    extutil.Ptr("JENKINS"),
			})
		}
	}
	return &action_kit_api.StopResult{
		Messages: &messages,
	}, nil
}
