# crane-hook

A Kubernetes cluster requirements checker for Blazemeter Private Locations. This tool verifies node resources, network connectivity, RBAC, and ingress configuration to ensure your cluster is ready to deploy Blazemeter workloads

## Features

- Checks node CPU, memory, and storage capacity
- Verifies network access to Blazemeter, Docker registry, and third-party endpoints
- Validates Kubernetes RBAC roles and bindings
- Confirms ingress or Istio gateway setup and TLS secret presence
- Designed to run as a Kubernetes Pod


## Usage

### [1.0] As a helm test hook

This image is packaged as a test hook with [`helm-crane`](https://github.com/Blazemeter/helm-crane/releases) chart, versioned `1.4.0` and later.
The test hook cal be executed by:
```sh
helm test <release-name>
```
It will automatically test the installation. 

**This is the only recommended method to execute the test hook. Running it manually or as an individual k8s pod may not produce the desired results.** 

See the [documentation](https://github.com/Blazemeter/helm-crane/blob/main/README.md) to learn more.


### [2.0] As a Kubernetes Pod

If you prefer, you can deploy the test image manually as a Kubernetes pod. However, you will need to configure the environment variables and roles required to test the installation correctly. 

Download the manifest YAML, see [`kubernetes/cranehook.yaml`](kubernetes/cranehook.yaml) for an example manifest. 


**Update the ENV Variables in the manifest file**, here are the details of what those ENV Variables mean: 

- `WORKING_NAMESPACE`: Namespace in which the crane and it's resources are installed.
- `ROLE_NAME`: Name of the Role that the crane serviceAccount is using
- `ROLE_BINDING_NAME`: Name of the RoleBinding that is binding the role to the serviceAccount used by crane.
- `SERVICE_ACCOUNT_NAME`: ServiceAccount used by crane, it is using the above roles for normal functioning of the crane/Blazemeter agent.
- `DOCKER_REGISTRY`: (Optional) Docker registry URL to check (it should be your private container image registry in case of image override)
- `KUBERNETES_WEB_EXPOSE_TYPE`: (Optional) INGRESS or ISTIO for Nginx or Istio type of ingress setup, required for Service virtualisation.
- `KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME`: (Optional) TLS secret name for nginx/istio type ingress setup
- `KUBERNETES_ISTIO_GATEWAY_NAME`: (Optional if using Istio) Gateway resource name
- `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`: (optional) Proxy settings, same as crane deployment manifest
**Update other manifest configurations**

- Update the namespace for all resources in the manifest to the correct one, that you are using to run crane. 
- In case of Istio based crane installation, please uncomment the istio related sections in the manifest file. 

Once the ENV variables are set, and configurations are updated, you can run the manifest yaml. 

```sh
kubectl apply -f cranehook.yaml 
```


## Output

- `[INFO]` messages indicate successful checks
- `[error]` messages indicate failed checks
- Exit code 0: all checks passed
- Exit code 1: one or more checks failed (the logs would list the failures/errors)
