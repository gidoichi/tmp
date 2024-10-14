# secrets-store-csi-driver-provider-infisical
[![Helm charts](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/secrets-store-csi-driver-provider-infisical&label=Helm+charts)](https://artifacthub.io/packages/search?repo=secrets-store-csi-driver-provider-infisical)

Unofficial Infisical provider for the Secret Store CSI Driver.

## Install
1. Prepare a Kubernetes Cluser running [Secret Store SCI Driver](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html)
1. Install Infisical secret proivder
   - If you can use [HELM](https://helm.sh/):
     ```
     helm repo add secrets-store-csi-driver-provider-infisical https://raw.githubusercontent.com/gidoichi/secrets-store-csi-driver-provider-infisical/main/charts
     helm install secrets-store-csi-driver-provider-infisical secrets-store-csi-driver-provider-infisical/secrets-store-csi-driver-provider-infisical
     ```
   - If you want to use kubectl (Using HELM is recommended, as some features are excluded from `./deployment`):
     ```
     kubectl apply -f ./deployment/infisical-csi-provider.yaml
     ```

## Usage
1. Create a new Infisical client using [Universal Auth](https://infisical.com/docs/documentation/platform/identities/universal-auth)
1. Store the Client ID and the Client Secret to a Kubernetes Secret as `client-id` key and `client-secret` key respectively
   ```
   # You can create a secret using the following command or applying `./examples/secret.yaml` after it is edited
   kubectl create secret generic infisical-secret-provider-auth-credentials --from-literal="client-id=$id" --from-literal="client-secret=$secret"
   ```
1. Create an SecretProviderClass referencing the secret
   ```
   # You should edit secretproviderclass.yaml to get secrets from provider
   kubectl apply -f ./examples/secretproviderclass.yaml
   ```
1. Create an Pod using the SecretProviderClass
   ```
   # This deployment lists and reads all secrets, then output logs of their contents
   kubectl apply -f ./examples/deployment.yaml
   ```

## Supported Features
Some features are not supported by this provider. Please refer to [this](https://secrets-store-csi-driver.sigs.k8s.io/providers#features-supported-by-current-providers) link for the list of features supported by the Secret Store CSI Driver.

| Features                            | Supported |
|-------------------------------------|-----------|
| [Sync as Kubernetes Secret][secret] | Yes       |
| [Rotation][rotation]                | No        |
| Windows                             | No        |
| Helm Chart                          | Yes       |

[secret]: https://secrets-store-csi-driver.sigs.k8s.io/topics/sync-as-kubernetes-secret
[rotation]: https://secrets-store-csi-driver.sigs.k8s.io/topics/secret-auto-rotation

### Test
The following are tested scenarios as part of CI. More detailed descriptions of these scenarios are available [here](https://github.com/kubernetes-sigs/secrets-store-csi-driver/tree/v1.4.5/test).

| Test Category                                          | Status                           |
|--------------------------------------------------------|----------------------------------|
| Mount tests                                            | [![mount-badge]][mount-ci]       |
| Sync as Kubernetes secrets                             | [![sync-badge]][sync-ci]         |
| Namespaced Scope SecretProviderClass                   | [![ns-badge]][ns-ci]             |
| Namespaced Scope SecretProviderClass negative test     | [![nsneg-badge]][nsneg-ci]       |
| Multiple SecretProviderClass                           | [![multiple-badge]][multiple-ci] |
| Autorotation of mount contents and Kubernetes secrets  | [![rotate-badge]][rotate-ci]     |
| Test filtered watch for `nodePublishSecretRef` feature | [![filtered-badge]][filtered-ci] |
| Windows tests                                          | [![windows-badge]][windows-ci]   |

[mount-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-mount.yml/badge.svg?branch=main
[mount-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-mount.yml?query=branch%3Amain
[sync-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-sync.yml/badge.svg?branch=main
[sync-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-sync.yml?query=branch%3Amain
[ns-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-namespaced.yml/badge.svg?branch=main
[ns-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-namespaced.yml?query=branch%3Amain
[nsneg-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-namespaced-neg.yml/badge.svg?branch=main
[nsneg-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-namespaced-neg.yml?query=branch%3Amain
[multiple-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-multiple.yml/badge.svg?branch=main
[multiple-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-multiple.yml?query=branch%3Amain
[rotate-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-rotate.yml/badge.svg?branch=main
[rotate-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-rotate.yml?query=branch%3Amain
[filtered-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-filtered.yml/badge.svg?branch=main
[filtered-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-filtered.yml?query=branch%3Amain
[windows-badge]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-windows.yml/badge.svg?branch=main
[windows-ci]: https://github.com/gidoichi/secrets-store-csi-driver-provider-infisical/actions/workflows/test-windows.yml?query=branch%3Amain
