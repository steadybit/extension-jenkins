package extjenkins

import (
	"context"
	"github.com/bndr/gojenkins"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-jenkins/config"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"time"
)

type jobDiscovery struct {
	jenkins *gojenkins.Jenkins
}

var (
	_ discovery_kit_sdk.TargetDescriber    = (*jobDiscovery)(nil)
	_ discovery_kit_sdk.AttributeDescriber = (*jobDiscovery)(nil)
)

func NewJobDiscovery(jenkins *gojenkins.Jenkins) discovery_kit_sdk.TargetDiscovery {
	discovery := &jobDiscovery{jenkins: jenkins}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 5*time.Minute),
	)
}

func (d *jobDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id: TargetTypeJob,
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1m"),
		},
	}
}

func (d *jobDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:      TargetTypeJob,
		Version: extbuild.GetSemverVersionStringOrUnknown(),
		Icon:    extutil.Ptr(TargetIconJob),

		Label: discovery_kit_api.PluralLabel{One: "Job", Other: "Jobs"},

		Category: extutil.Ptr("Jenkins"),

		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "jenkins.job.name.full.display"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "jenkins.job.name.full.display",
					Direction: "ASC",
				},
			},
		},
	}
}

func (d *jobDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	return []discovery_kit_api.AttributeDescription{
		{
			Attribute: "jenkins.job.name",
			Label: discovery_kit_api.PluralLabel{
				One:   "Job name",
				Other: "Job names",
			},
		},
		{
			Attribute: "jenkins.job.name.full",
			Label: discovery_kit_api.PluralLabel{
				One:   "Job full name",
				Other: "Job full names",
			},
		},
		{
			Attribute: "jenkins.job.name.full.display",
			Label: discovery_kit_api.PluralLabel{
				One:   "Job display name",
				Other: "Job display names",
			},
		},
		{
			Attribute: "jenkins.job.url",
			Label: discovery_kit_api.PluralLabel{
				One:   "Job url",
				Other: "Job urls",
			},
		},
	}
}

func (d *jobDiscovery) DiscoverTargets(ctx context.Context) ([]discovery_kit_api.Target, error) {
	jobs, err := getAllJobsRecursive(ctx, d.jenkins)
	if err != nil {
		return nil, extension_kit.ToError("Failed to fetch jobs.", err)
	}

	targets := make([]discovery_kit_api.Target, len(jobs))
	for i, job := range jobs {
		targets[i] = discovery_kit_api.Target{
			Id:         job.Base,
			TargetType: TargetTypeJob,
			Label:      job.GetDetails().FullDisplayName,
			Attributes: map[string][]string{
				"jenkins.job.name":              {job.GetName()},
				"jenkins.job.name.full":         {job.GetDetails().FullName},
				"jenkins.job.name.full.display": {job.GetDetails().FullDisplayName},
				"jenkins.job.url":               {job.GetDetails().URL},
				"jenkins.job.class":             {job.GetDetails().Class},
			},
		}

		var parameters []gojenkins.ParameterDefinition
		for _, property := range job.Raw.Property {
			parameters = append(parameters, property.ParameterDefinitions...)
		}
		if len(parameters) > 0 {
			parameterAttribute := make([]string, len(parameters))
			for j, parameter := range parameters {
				parameterAttribute[j] = parameter.Name
			}
			targets[i].Attributes["jenkins.job.parameter"] = parameterAttribute
		}
	}
	return discovery_kit_commons.ApplyAttributeExcludes(targets, config.Config.DiscoveryAttributesExcludesJob), nil
}

func getAllJobsRecursive(ctx context.Context, jenkins *gojenkins.Jenkins) ([]*gojenkins.Job, error) {
	var allJobs []*gojenkins.Job

	jobs, err := jenkins.GetAllJobs(ctx)
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.Raw.Class == "com.cloudbees.hudson.plugins.folder.Folder" {
			folderJobs, err := getAllJobsInFolderRecursive(ctx, jenkins, job.GetDetails().Jobs, job.GetName())
			if err != nil {
				return nil, err
			}
			allJobs = append(allJobs, folderJobs...)
		} else {
			allJobs = append(allJobs, job)
		}
	}
	return allJobs, nil
}

func getAllJobsInFolderRecursive(ctx context.Context, jenkins *gojenkins.Jenkins, innerJobs []gojenkins.InnerJob, parentIDs ...string) ([]*gojenkins.Job, error) {
	var allJobs []*gojenkins.Job

	for _, innerJob := range innerJobs {
		var job, err = jenkins.GetJob(ctx, innerJob.Name, parentIDs...)
		if err != nil {
			return nil, err
		}

		if job.Raw.Class == "com.cloudbees.hudson.plugins.folder.Folder" {

			subJobs, err := getAllJobsInFolderRecursive(ctx, jenkins, job.GetDetails().Jobs, append(parentIDs, innerJob.Name)...)
			if err != nil {
				return nil, err
			}
			allJobs = append(allJobs, subJobs...)
		} else {
			allJobs = append(allJobs, job)
		}
	}
	return allJobs, nil
}
