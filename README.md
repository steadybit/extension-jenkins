<img src="./jenkins.svg" height="180" align="right" alt="Jenkins Logo">

# Steadybit extension-jenkins

A [Steadybit](https://www.steadybit.com/) extension for [Jenkins](https://www.jenkins.io/).

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_jenkins).


## Configuration

| Environment Variable          | Helm value         | Meaning                                                                 | Required | Default |
|-------------------------------|--------------------|-------------------------------------------------------------------------|----------|---------|
| STEADYBIT_EXTENSION_BASE_URL  | `jenkins.baseUrl`  | The base URL of your Jenkins installation, like 'https://ci.jenkins.io' | yes      |         |
| STEADYBIT_EXTENSION_API_USER  | `jenkins.apiUser`  | The Jenkins API User                                                    | yes      |         |
| STEADYBIT_EXTENSION_API_TOKEN | `jenkins.apiToken` | The Jenkins API Token                                                   | yes      |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-jenkins.enabled=true \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-jenkins/blob/main/charts/steadybit-extension-jenkins/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-jenkins/blob/main/charts/steadybit-extension-jenkins).

```bash
helm repo add steadybit-extension-jenkins https://steadybit.github.io/extension-jenkins
helm repo update
helm upgrade steadybit-extension-jenkins \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-agent \
    --set jenkins.baseURL=<YOUR_BASE_URL> \
    --set jenkins.apiUser=<YOUR_API_USER> \
    --set jenkins.apiToken=<YOUR_API_TOKEN> \
    steadybit-extension-jenkins/steadybit-extension-jenkins
```

### Linux Package

Please use
our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts)
to install the extension on your Linux machine. The script will download the latest version of the extension and install
it using the package manager.

After installing, configure the extension by editing `/etc/steadybit/extension-jenkins` and then restart the service.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-registration) for more
information about extension registration and how to verify.

## Version and Revision

The version and revision of the extension:
- are printed during the startup of the extension
- are added as a Docker label to the image
- are available via the `version.txt`/`revision.txt` files in the root of the image
