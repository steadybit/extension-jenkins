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
		Icon:    extutil.Ptr("data:image/svg+xml,%3Csvg%20viewBox%3D%220%200%2024%2024%22%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%3E%3Cpath%20d%3D%22m2.873%2024h-.975a3.866%203.866%200%200%201%20-.07-.197c-.215-.666-.594-1.49-.692-2.154-.146-.984.78-1.039%201.374-1.465.915-.66%201.635-1.025%202.627-1.621.295-.178%201.182-.623%201.281-.828.201-.408-.345-.982-.49-1.301-.225-.506-.345-.936-.376-1.434-.824-.131-1.455-.627-1.844-1.185-.63-.925-1.066-2.635-.525-3.936.045-.103.254-.305.285-.463.06-.309-.105-.72-.121-1.049-.06-1.692.285-3.15%201.426-3.66.463-1.84%202.113-2.453%203.673-3.367.581-.342%201.224-.562%201.89-.807%202.372-.877%206.028-.712%207.995.783.836.633%202.176%201.971%202.656%202.939%201.262%202.555%201.17%206.826.287%209.935-.121.421-.291%201.032-.533%201.533-.168.349-.689%201.05-.625%201.36.064.314%201.189%201.17%201.432%201.395.434.422%201.26.975%201.324%201.5.07.557-.248%201.336-.41%201.875-.217.721-.436%201.441-.654%202.131h-18.937zm11.104-3.541c-.545-.299-1.361-.621-2.065-.756-.869-.164-.779%201.188-.75%201.994.03.643.361%201.316.511%201.744.075.197.089.41.255.449.3.068%201.29-.326%201.575-.479.6-.328%201.064-.844%201.574-1.189.016-.17.016-.34.031-.508a2.648%202.648%200%200%200%20-1.096-.277c.314-.15.75-.15%201.035-.332l.016-.193c-.496-.031-.689-.254-1.021-.436zm7.455%202.936c.193-.619.359-1.221.465-1.752.059-.287.215-.918.178-1.176-.059-.459-.684-.799-1.004-1.086-.584-.525-.951-.975-1.561-1.469-.248.375-.779.615-.982.914%201.447-.689%201.709%202.625%201.141%203.689.09.33.391.451.514.736l-.086.166h1.289c.014%200%20.031%200%20.045.014zm-6.635-.012c-.049-.074-.1-.135-.15-.209l-.301.195h.451zm2.771%200c.008-.209.018-.404.029-.598-.529.029-.824-.48-1.195-.527-.324-.045-.6.361-1.02.195-.096.105-.184.227-.285.316.154.18.295.375.424.584h.816c.014-.164.135-.285.299-.285.166%200%20.285.121.285.27h.66zm2.116%200c-.314-.479-.947-.898-1.68-.555l-.031.541h1.711zm-8.51%200-.105-.344c-.225-.721-.359-1.26-.405-1.68-.914-.436-1.875-.871-2.654-1.426-.15-.105-1.109-1.35-1.23-1.305-1.739.676-3.359%201.859-4.814%202.984.256.557.48%201.141.69%201.74h8.505zm8.265-2.113c-.029-.512-.164-1.561-.48-1.74-.66-.391-1.846.779-2.34.943.045.15.135.271.15.48.285-.074.645-.029.898.092-.299.029-.629.029-.824.164-.074.195.016.48-.029.764.689.197%201.5.303%202.385.332.164-.227.225-.645.211-1.082zm-4.08-.36c-.045.375.045.51.119.943%201.26.391%201.035-1.74-.135-.959zm-6.598-1.41c-.45.457%201.271%201.082%201.814%201.115%200-.291.165-.564.135-.771-.649-.117-1.502-.041-1.945-.346zm5.565.215c0%20.043-.061.029-.068.064.58.451%201.014.545%201.803.51.354-.262.67-.563%201.043-.807-.855.074-1.932.607-2.775.229zm3.42-17.727c-1.605-.906-4.35-1.591-6.076-.731-1.38.692-3.27%201.841-3.899%203.292.6%201.402-.166%202.686-.226%204.109-.018.758.36%201.42.391%202.243-.2.338-.825.38-1.26.356-.146-.729-.401-1.549-1.155-1.631-1.064-.115-1.845.765-1.891%201.684-.06%201.079.834%202.864%202.086%202.745.488-.046.608-.541%201.139-.541.285.57-.445.75-.523%201.155-.016.105.059.511.104.705.233.944.744%202.159%201.245%202.88.635.9%201.884%201.051%203.229%201.141.24-.525%201.125-.48%201.706-.346-.691-.27-1.336-.945-1.875-1.529-.615-.676-1.23-1.41-1.261-2.28%201.155%201.604%202.1%202.999%204.201%203.704%201.59.525%203.449-.254%204.664-1.109.51-.359.811-.93%201.17-1.439%201.35-1.936%201.98-4.71%201.846-7.395-.061-1.111-.061-2.221-.436-2.955-.389-.781-1.695-1.471-2.475-.781-.15-.764.629-1.229%201.545-.959-.66-.855-1.336-1.859-2.266-2.385l.017.067zm-4.164%2012.908c.615%201.544%202.725%201.363%204.506%201.323-.084.194-.256.435-.465.515-.57.232-2.146.408-2.938-.012-.506-.271-.824-.873-1.102-1.227-.137-.172-.795-.608-.012-.609zm.164-.871c.893.465%202.52.518%203.732.48.066.268.066.594.068.914-1.551.08-3.387-.304-3.795-1.395h-.005zm6.676-.585c-.473.9-1.145%201.897-2.539%201.928-.023-.284-.045-.735%200-.904%201.064-.103%201.727-.646%202.543-1.017zm-.649-.667c-1.02.66-2.154%201.375-3.824%201.21-.352-.31-.486-1.001-.141-1.458.182.313.061.885.57.969.945.166%202.039-.578%202.73-.84.42-.712-.045-.975-.42-1.432-.781-.931-1.83-2.1-1.801-3.51.314-.225.346.345.391.449.404.961%201.424%202.176%202.174%203%20.18.211.48.391.51.525.092.389-.254.854-.209%201.109zm-13.44-.675c-.314-.184-.393-.99-.768-1.01-.535-.03-.438%201.049-.436%201.679-.37-.329-.435-1.364-.164-1.89-.308-.15-.445.165-.618.285.22-1.59%202.34-.734%201.989.96zm-1.619-6.104c-.685.756-.54%202.174-.459%203.188%201.244-.785%202.898.06%202.883%201.395.595-.016.223-.744.115-1.215-.353-1.529.592-3.188.041-4.59-1.064.083-1.939.519-2.578%201.214zm9.12%201.113c.307.562.404%201.148.84%201.57.195.191.574.424.387.951-.045.121-.365.391-.551.45-.674.195-2.254.03-1.721-.81.563.015%201.314.359%201.732-.045-.314-.525-.885-1.53-.674-2.13zm6.199-.013h.068c.33.668.6%201.375%201.004%201.965-.27.629-2.053%201.19-2.023.057.389-.17%201.049-.035%201.395-.25-.193-.556-.48-1.006-.434-1.771zm-6.928-1.617c-1.422-.33-2.131.592-2.56%201.553-.384-.094-.231-.615-.135-.883.255-.701%201.28-1.633%202.119-1.506.359.057.848.386.576.834zm-3.462-3.885c-1.56.44-3.56%201.574-4.2%202.974.495-.07.84-.321%201.33-.351.186-.016.428.074.641.015.424-.104.78-1.065%201.102-1.409.311-.346.685-.497.941-.811.166-.09.408-.074.42-.33-.074-.075-.15-.135-.233-.105v.017z%22%2F%3E%3C%2Fsvg%3E"),

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
