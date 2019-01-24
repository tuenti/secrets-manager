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

- [vault-crd](https://github.com/DaspawnW/vault-crd). This is the tool that really inspired *secrets-manager*. We opened this [issue](https://github.com/DaspawnW/vault-crd/issues/4) asking for token renewal or other login mechanism. While the author is very responsive answering, we could not wait for an implementation and since we were more used to Go than Java we decided to write *secrets-manager*. We are very thankful to the author of *vault-crd*, since has been really inspiring. Some differences:
  - *vault-crd* uses Hashicorp Vault as the source of truth, while *secrets-manager* has been designed to support other backends (we only support Vault for now,though).
  - *vault-crd* uses [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) while *secrets-manager* uses configmaps. Configmap was the very first step, but we will migrate it to CRDs as part of our short-term roadmap.
  - *vault-crd* supports KV1 and pki engines, while *secrets-manager* supports KV1 and KV2. It is also in our roadmap to support more engines.

# How it works

*secrets-manager* gets initialized with a Vault token and a Kubernetes configmap. While it's running it will be checking in the background:

- If Vault token is close to expire and if that's the case, renewing it.
- The Kubernetes configmap data, reloading the mounted config file in case there is any change.


## Configmap

*secrets-manager* configmap looks like the following example:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: secrets-manager-configmap
  namespace: secrets-manager
data:
  secretDefinitions: |-
    - data:
        tls.crt:
          encoding: base64
          key: crt
          path: secret/data/tls-example-io
        tls.key:
          encoding: base64
          path: secret/data/tls-example-io
          key: key
      name: tls-example-io
      namespaces:
      - example
      type: kubernetes.io/tls
    - data:
        dbuser:
          key: user
          path: secret/data/db-credentials
        dbpassword:
          key: password
          path: secret/data/db-credentials
      name: db-credentials
      namespaces:
      - webapp
      - foo
      - bar
      type: Opaque
```

### Secrets Definition

- `name`: This will be the name of the secret created in Kubernetes.
- `namespaces`: A list of namespaces where the secret has to be created.
- `type`: Kubernetes secret type. One of `kubernetes.io/tls`, `Opaque`.
- `data`: This will contain the Kubernetes secret data keys as a map of datasources. Each datasource will contain the way to access the secret in the secret backend source of truth, via a `path` and `key`. And optional `encoding` key can be provided if your secrets are stored in `base64`. The absence of `encoding` or `encoding: text` means no encoding.

**NOTE**: We let the user all the responsibility to set the whole Vault path. So it is important to know which path a secret engine needs to be set. For instance, with the KV version 1 all secrets are stored in `secret/` whereas with the KV version 2, all secrets go under `secret/data/`

## Flags

| Flag | Default | Description |
| ------ | ------- | ------ |
| `log.level` | warn | Minimum log level |
| `log.format` | text | Log format, one of text or json |
| `backend`| vault | Selected backend. Only vault supported for now |
| `config.backend-timeout`| 5s | Backend connection timeout |
| `config.backend-scrape-interval`| 15s | Scraping secrets from backend interval |
| `config.config-map`| 15s | Name of the configmap with *secrets-manager* settings (format: `namespace/name`)  (default "secrets-manager-config") |
| `config.configmap-refresh-interval`| 15s | ConfigMap refresh interval |
| `vault.url` | https://127.0.0.1:8200 | Vault address. `VAULT_ADDR` environment would take precedence. |
| `vault.token` | `""` | Vault token. `VAULT_TOKEN` environment would take precedence. |
| `vault.engine` | kv2 | Vault secrets engine to use. Only key/value engines supported. Default is kv version 2 |
| `vault.max-token-ttl` | 300 |Max seconds to consider a token expired. |
| `vault.token-polling-period` | 15s | Polling interval to check token expiration time. |
| `vault.renew-ttl-increment` | 600 | TTL time for renewed token. |
| `listen-address` | `:8080` | The address to listen on for HTTP requests. |

## Prometheus Metrics

`secrets-manager` exposes the following [Prometheus](https://prometheus.io) metrics at `http://$cfg.listen-addr/metrics`:

| Metric| Type| Description| Labels|
| ------| ----|------------| ------|
|`secrets_manager_vault_token_expired` | Gauge | Whether or not token TTL is under `vault.max-token-ttl`: 1 = expired; 0 = still valid | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name"` |
|`secrets_manager_vault_token_ttl` | Gauge | Vault token TTL | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name"` |
|`secrets_manager_vault_token_lookup_errors_count`| Counter | Vault token lookup-self errors counter | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name", "error"` |
|`secrets_manager_vault_token_renew_errors_count`| Counter | Vault token renew-self errors counter | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name", "error"` |
|`secrets_manager_read_secret_errors_count`| Counter | Vault read operations counter | `"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name", "path", "key", "error"` |
| `secrets_manager_secret_sync_errors_count`| Counter |Secrets sync error counter|`"name", "namespace"`|
|`secrets_manager_secret_last_updated`| Gauge |The last update timestamp as a Unix time (the number of seconds elapsed since January 1, 1970 UTC)|`"name", "namespace"`|

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

To create a regular token attached to a policy:

`$ vault token create -ttl=1h -policy=my-policy`

To create a periodic token you can create a token role as follows:

`$ vault write auth/token/roles/secrets-manager allowed_policies=my-policy period=1h`

And then generate a token associated to that role

`$ vault token create -role="secrets-manager`

## Deployment
*secrets-manager* has been designed to be deployed in Kubernetes as it reads its config file from Kubernetes Configmap. Future versions of *secrets-manager* may use Custom Resource Definitions instead. You will find a full deployment example in the [examples/](examples) folder.

## Credits & Contact

*secrets-manager* is developed and maintained by [Tuenti Technologies S.L.](http://github.com/tuenti)

You can follow Tuenti engineering team on Twitter [@tuentieng](http://twitter.com/tuentieng).

## License

*secrets-manager* is available under the Apache License, Version 2.0. See LICENSE file
for more info.
