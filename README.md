## WARNING: Latest version of `secrets-manager` introduces  two major breaking changes:
  - Vault token based authentication will no longer be supported in favor of AppRole authentication. 
  - Configmap-based secret definitions will be deprecated in favor of CRDs.
Checkout `CHANGELOG.md` for more information.

# secrets-manager 
[![CircleCI](https://circleci.com/gh/tuenti/secrets-manager/tree/master.svg?style=svg)](https://circleci.com/gh/tuenti/secrets-manager/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/tuenti/secrets-manager)](https://goreportcard.com/report/github.com/tuenti/secrets-manager)
[![codecov](https://codecov.io/gh/tuenti/secrets-manager/branch/master/graph/badge.svg)](https://codecov.io/gh/tuenti/secrets-manager)

A tool to keep your Kubernetes secrets in sync with Vault

# Rationale

Lots of companies use [Vault](https://www.vaultproject.io) as their secrets store backend for multiple kind of secrets and different purposes. Kubernetes brings a nice secrets API, but it means that you have two different sources of truth for your secrets.

*secrets-manager* tries to solve this, by reading secrets from Vault and comparing them to Kubernetes secrets, creating and updating them as you do it in Vault.

# How does it compare to other tools?

- [cert-manager](https://github.com/jetstack/cert-manager). *cert-manager* solves a different issue, automation around issuing and renewing certificates. It integrates with Let's Encrypt and Vault (using the pki backend) being those the certificates issuer. While this is really powerful and really a tool which is fully compatible with *secrets-manager*, it does not really sync a secret from a secret backend. *secrets-manager* is a more generic tool where you can sync certificates or any kind of secret from the source of truth of your secrets to Kubernetes secrets.

- [vault-operator](https://github.com/coreos/vault-operator). This manages vault clusters in Kubernetes, so it is a completely different tool.

- [vault-crd](https://github.com/DaspawnW/vault-crd). This is the tool that really inspired *secrets-manager*. We opened this [issue](https://github.com/DaspawnW/vault-crd/issues/4) asking for token renewal or other login mechanism. While the author is very responsive answering, we could not wait for an implementation and since we were are more familiar with Go than Java we decided to write *secrets-manager*. We are very thankful to the author of *vault-crd*, since it has been really inspiring. Some differences:
  - *vault-crd* uses Hashicorp Vault as the source of truth, while *secrets-manager* has been designed to support other backends (we only support Vault for now,though).
  - *vault-crd* supports KV1 and pki engines, while *secrets-manager* supports KV1 and KV2. It is also in our roadmap to support more engines.

# How it works

*secrets-manager* will login to Vault using AppRole credentials and it will start a reconciliation loop watching for changes in `SecretsDefinition` objects. In background it will run two main operations: 

- If Vault token is close to expire and if that's the case, renewing it. If it can't renew, it will try to re-login.
- It will re-queue `SecretsDefinition` events and in every event loop it will verify if the current Kubernetes secret it is in the desired state by comparing it with the data in Vault and creating/updating them accordingly

## Custom Resource Definition (CRD)

*secrets-manager* now uses [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions) to extend Kubernetes APIs with a new `SecretDefinition` object that it will watch.

To install the CRD in your cluster: `kubectl apply -f crd.yaml`


### Secrets Definition

- `name`: This will be the name of the secret created in Kubernetes.
- `type`: Kubernetes secret type. One of `kubernetes.io/tls`, `Opaque`.
- `keysMap`: This will contain the Kubernetes secret data keys as a map of datasources. Each datasource will contain the way to access the secret in the secret backend source of truth, via a `path` and  a `key`. And optional `encoding` key can be provided if your secrets are codified in `base64`. The absence of `encoding` or `encoding: text` means no encoding.

**NOTE**: We let the user all the responsibility to set the whole Vault path. So it is important to know which path a secret engine needs to be set. For instance, with the KV version 1 all secrets are stored in `secret/` whereas with the KV version 2, all secrets go under `secret/data/`

An example of a `secretdefinition` object

```
$ cat > secretdefinition-sample.yaml <<EOF
---
apiVersion: secrets-manager.tuenti.io/v1alpha1
kind: SecretDefinition
metadata:
  name: secretdefinition-sample
spec:
  # Add fields here
  name: supersecretnew
  keysMap:
    decoded:
      path: secret/data/pathtosecret1
      encoding: base64
      key: value
    raw:
      path: secret/data/pathtosecret1
      key: value

EOF
```

To deploy it just run `kubectl apply -f secretdefinition-sample.yaml`
## Flags

| Flag | Default | Description |
| ------ | ------- | ------ |
| `backend`| vault | Selected backend. Only vault supported for now |
| `enable-debug-log` | `false` | Enable this to get more logs verbosity and debug messages.|
| `enable-leader-election` | `false` | Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.|
| `reconcile-period`| 5s | How often the controller will re-queue secretdefinition events |
| `config.backend-timeout`| 5s | Backend connection timeout |
| `vault.url` | https://127.0.0.1:8200 | Vault address. `VAULT_ADDR` environment would take precedence. |
| `vault.role-id` | `""` | Vault appRole `role_id`. `VAULT_ROLE_ID` environment would take precedence. |
| `vault.secret-id` | `""` | Vault appRole `secret_id`. `VAULT_SECRET_ID` environment would take precedence. |
| `vault.engine` | kv2 | Vault secrets engine to use. Only key/value engines supported. Default is kv version 2 |
| `vault.max-token-ttl` | 300 |Max seconds to consider a token expired. |
| `vault.token-polling-period` | 15s | Polling interval to check token expiration time. |
| `vault.renew-ttl-increment` | 600 | TTL time for renewed token. |
| `metrics-addr` | `:8080` | The address to listen on for HTTP requests. |
| `watch-namespaces` | `""` | Comma separated list of namespaces that secrets-manager will watch for `SecretDefinitions`. By default all namespaces are watched. |
| `exclude-namespaces` | `""` | Comma separated list of namespaces that secrets-manager will not watch for `SecretDefinitions`. By default all namespaces are watched. |

## Prometheus Metrics

`secrets-manager` exposes the following [Prometheus](https://prometheus.io) metrics at `http://$cfg.listen-addr/metrics`:

| Metric| Type| Description| Labels|
| ------| ----|------------| ------|
|`secrets_manager_vault_max_token_ttl` | Gauge | `secrets-manager` max Vault token TTL | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name"` |
|`secrets_manager_vault_token_ttl` | Gauge | Vault token TTL | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name"` |
|`secrets_manager_vault_token_renewal_errors_total`| Counter | Vault token renewal errors counter | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name", "vault_operation", "error"` |
|`secrets_manager_controller_secret_read_errors_total`| Counter | Errors total count when reading a secret from Kubernetes | `"name", "namespace"` |
| `secrets_manager_controller_sync_errors_total`| Counter |Secrets synchronization total errors.|`"name", "namespace"`|
|`secrets_manager_controller_last_sync_status`| Gauge |The result of the last sync of a secret. 1 = OK, 0 = Error|`"name", "namespace"`|

## Getting Started with Vault

### Vault Policies

We do recommend you use policies to make sure you grant `secrets-manager` only to those secrets you need available in your Kubernetes cluster. An example of simple policy could be:

```
path "secret/data/my-k8s-cluster/*" {
  capabilities = ["read"]
}
```

To create this policy:

```
$ cat > my-policy.hcl <<EOF
path "secret/data/my-k8s-cluster/*" {
  capabilities = ["read"]
}
EOF

$ cat my-policy.hcl | vault policy write my-policy -
```

### Vault Tokens

Vault tokens will be renewed by `secrets-manager` if the `ttl` is lower than `vault.max-token-ttl` and the token is renewable. But as per Vault's [documentation](https://www.vaultproject.io/docs/concepts/tokens.html#the-general-case), regular tokens will have their own max TTL that it's calculated on every renewal, so that a token will eventually expire. This can be ok for your use case, but for others a [periodic token](https://www.vaultproject.io/docs/concepts/tokens.html#periodic-tokens) could be much more convinient. In the case of a periodic token, the `period` will invalidate the `vault.renew-ttl-increment` option.


### Vault AppRole
Vault token as a login mechanism has been deprecated in favor of the [AppRole](https://www.vaultproject.io/docs/auth/approle.html) authentication method for `secrets-manager`. 
`secrets-manager` will still renew the token obtained after login in, but will make `secrets-manager` more resilient in case of a token has expired due to network issues, Vault sealed, etc.

So instead of expecting a token, `secrets-manager` expects a `role_id` and a `secret_id` to connect to Vault.

To create a role with a permanent `secret_id` attached to a policy:

`$ vault write auth/approle/role/secrets-manager policies=my-policy secret_id_num_uses=0 secret_id_ttl=0`

To get a `secret_id`:

`$ vault write -force auth/approle/role/secrets-manager/secret-id`

To get the `role_id`:

`$ vault read auth/approle/role/secrets-manager/role-id`

## Deployment
*secrets-manager* has been designed to be deployed in Kubernetes, you will find a full deployment example in the [config/samples](config/samples) folder.

## Credits & Contact

*secrets-manager* is developed and maintained by [Tuenti Technologies S.L.](http://github.com/tuenti)

You can follow Tuenti engineering team on Twitter [@tuentieng](http://twitter.com/tuentieng).

## License

*secrets-manager* is available under the Apache License, Version 2.0. See LICENSE file
for more info.
