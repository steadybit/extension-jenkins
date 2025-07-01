<img src="./jenkins.svg" height="180" align="right" alt="Jenkins Logo">

# Steadybit extension-jenkins

A [Steadybit](https://www.steadybit.com/) extension for [Jenkins](https://www.jenkins.io/).

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_jenkins).


## Configuration

| Environment Variable                          | Helm value         | Meaning                                                                 | Required | Default |
|-----------------------------------------------|--------------------|-------------------------------------------------------------------------|----------|---------|
| STEADYBIT_EXTENSION_BASE_URL                  | `jenkins.baseUrl`  | The base URL of your Jenkins installation, like 'https://ci.jenkins.io' | yes      |         |
| STEADYBIT_EXTENSION_API_USER                  | `jenkins.apiUser`  | The Jenkins API User                                                    | yes      |         |
| STEADYBIT_EXTENSION_API_TOKEN                 | `jenkins.apiToken` | The Jenkins API Token                                                   | yes      |         |
| STEADYBIT_EXTENSION_JOB_START_TIMEOUT_SECONDS |                    | Timeout for a job to start, otherwise an error is returned              | yes      | 60      |

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
--set extension-jenkins.jenkins.baseURL=<YOUR_BASE_URL> \
--set extension-jenkins.jenkins.apiUser=<YOUR_API_USER> \
--set extension-jenkins.jenkins.apiToken=<YOUR_API_TOKEN> \
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

## Importing your own certificates

You may want to import your own certificates for connecting to Jenkins instances with self-signed certificates. This can be done in two ways:

### Option 1: Using InsecureSkipVerify

The extension provides the `insecureSkipVerify` option which disables TLS certificate verification. This is suitable for testing but not recommended for production environments.

```yaml
jenkins:
  insecureSkipVerify: true
```

### Option 2: Mounting custom certificates

Mount a volume with your custom certificates and reference it in `extraVolumeMounts` and `extraVolumes` in the helm chart.

This example uses a config map to store the `*.crt`-files:

```shell
kubectl create configmap -n steadybit-agent jenkins-self-signed-ca --from-file=./self-signed-ca.crt
```

```yaml
extraVolumeMounts:
  - name: extra-certs
    mountPath: /etc/ssl/extra-certs
    readOnly: true
extraVolumes:
  - name: extra-certs
    configMap:
      name: jenkins-self-signed-ca
extraEnv:
  - name: SSL_CERT_DIR
    value: /etc/ssl/extra-certs:/etc/ssl/certs
```

## Version and Revision

The version and revision of the extension:
- are printed during the startup of the extension
- are added as a Docker label to the image
- are available via the `version.txt`/`revision.txt` files in the root of the image
