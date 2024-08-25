# secrets-store-csi-driver-provider-infisical
[![Helm charts](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/secrets-store-csi-driver-provider-infisical)](https://artifacthub.io/packages/search?repo=secrets-store-csi-driver-provider-infisical&label=Helm+charts)

Unofficial Infisical provider for the Secret Store CSI Driver.

## Install
- Prepare a Kubernetes Cluser running [Secret Store SCI Driver](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html)
- Create a new Infisical client using [Universal Auth](https://infisical.com/docs/documentation/platform/identities/universal-auth)
- Store the Client ID and the Client Secret to a Kubernetes Secret as `client-id` key and `client-secret` key respectively  
  ```
  kubectl create secret generic infisical-secret-provider-auth-credentials --from-literal="client-id=$id" --from-literal="client-secret=$secret"
  ```
- Install Infisical secret proivder
  ```
  kubectl apply -f ./manifests.yaml
  ```
