# crane-testhook

A Kubernetes cluster requirements checker for Blazemeter Private Locations. This tool verifies node resources, network connectivity, RBAC, and ingress configuration to ensure your cluster is ready to deploy Blazemeter workloads

## Features

- Checks node CPU, memory, and storage capacity
- Verifies network access to Blazemeter, Docker registry, and third-party endpoints
- Validates Kubernetes RBAC roles and bindings
- Confirms ingress or Istio gateway setup and TLS secret presence
- Designed to run as a Kubernetes Pod


## Usage

### As a helm test hook

This image is packaged as a test hook with [`helm-crane`](https://github.com/Blazemeter/helm-crane/releases) chart, versioned `1.4.0` and later.
The test hook cal be executed by:
```sh
helm test <release-name>
```
It will automatically test the installation. 

**This is the only recommended method to execute the test hook. Running it manually or as an individual k8s pod may not produce the desired results.** 

See the [documentation](https://github.com/Blazemeter/helm-crane/blob/main/README.md) to learn more.


### As a Kubernetes Pod

See [`kubernetes/cranehook.yaml`](kubernetes/cranehook.yaml) for an example manifest. 

Apply it with:

```sh
kubectl apply -f kubernetes/cranehook.yaml
```

The pod will run the checks and exit with code 0 if all requirements are met, or 1 if any check fails.

**NOTE**
- If you deploy crane-hook manually, with `kubectl apply` method, make sure the environment variables are set correctly as per the type of crane installation. See Environment Variables below.


## Environment Variables

- `WORKING_NAMESPACE`: Namespace in which the crane and it's resources are installed. 
- `ROLE_NAME`: Name of the Role to check
- `ROLE_BINDING_NAME`: Name of the RoleBinding to check
- `SERVICE_ACCOUNT_NAME`: ServiceAccount to check, which is using the above roles
- `KUBERNETES_WEB_EXPOSE_TYPE`: `INGRESS` or `ISTIO` for *Nginx* or *Istio* type of ingress setup, required for Service virtualisation.
- `DOCKER_REGISTRY`: Docker registry URL to check (it should be "gcr.io/verdant-bulwark-278")
- `KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME`: TLS secret name for nginx/istio type ingress setup
- `KUBERNETES_ISTIO_GATEWAY_NAME`: (if using Istio) Gateway resource name
- `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`: (optional) Proxy settings, same as crane deployment manifest


## Output

- `[INFO]` messages indicate successful checks
- `[error]` messages indicate failed checks
- Exit code 0: all checks passed
- Exit code 1: one or more checks failed (the logs would list the failures/errors)
