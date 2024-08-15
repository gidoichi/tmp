# secrets-store-csi-driver-provider-infisical
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

## Usage

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: test-csi-provider
  namespace: default
spec:
  provider: infisical
  parameters:
    projectSlug: example-project-em-9-e
    envSlug: dev
    secretsPath: /
    # Kubernetes Secret name storing Infisical client ID and client secret
    authSecretName: infisical-secret-provider-auth-credentials
    # Kubernetes Secret namespace
    authSecretNamespace: default
  secretObjects:
  - secretName: test-csi-provider
    type: Opaque
    data:
    - objectName: DATABASE_URL
      key: url
```
