## Unreleased

- [FEATURE] Add support for Azure KeyVault backend

## v2.0.1 2022-04-04

- [BUG] Fix nil pointer dereference bug in controller's regular kubernetes client

## v2.0.0 2022-02-21

- [FEATURE] Populating Labels and Annotations from the SecretDefinition to the generated Secret.
- [ENHANCEMENT] Updates the `managed-by` and `updatedAt` labels to more closely match k8s recommended values (using annotations and recommended labels), as seen below:
```yaml
annotations:
    secrets-manager.tuenti.io/lastUpdateTime: 2020-04-22T14.34.17Z
labels:
   app.kubernetes.io/managed-by: secrets-manager
```
- [ENHANCEMENT] Update to kubebuilder 3.1.0

## v1.1.0 2021-01-05

- [BEHAVIOUR] Using flags watch-namespaces / exclude-namespaces. They interact differently.
  - All namespaces are watched. A namespace is excluded if it is specified within the *exclude-namespaces* flag.
- [FEATURE] Adding **auth-method** param to specify Vault authentication method.
  - Adding vault authentication method from kubernetes. With **auth-method** param set to **kubernetes**.
- [BUG] set the controller name to something unique avoid 'duplicate metrics collector registration attempted' errors.

## no code related changes 2020-04-28

- No logic changes in secrets-manager. But we are going to stablish some changes in the project management:
  - Now versions are going to follow [semantic versioning](https://semver.org/) where version tags are going to have the 'v' preffix, they are going to be just:
    - v{major}.{minor}.{patch}, where major, minor and path are integers

  - From now on we are going to push release candidates to the [docker registry](https://hub.docker.com/repository/docker/tuentitech/secrets-manager)

## v1.0.2 2019-11-17

Stable release. Adds watching specific namespaces (see v1.0.2-rc.1) and some minor fixes.

### Fixes
- [#47 missing return provokes wrong metrics delivery](https://github.com/tuenti/secrets-manager/issues/47)
- [#37 Unable to build 1.0.1](https://github.com/tuenti/secrets-manager/issues/37)

## v1.0.2-rc.1 2019-09-30

### Fixes
- [#38 add the ability to watch secretDefinitions scoped to a particular namespace](https://github.com/tuenti/secrets-manager/issues/38)

## v1.0.1 2019-08-14
### Fixes
- Deleting a `SecretDefinition` hangs if the corresponding secret does not exist.
- Invalid metric names in README

### Deprecates
- Unused prometheus metrics `secrets_manager_controller_update_secret_errors_total` and `secrets_manager_controller_last_updated`

## v1.0.0 2019-07-29
Stable release

## v1.0.0-rc.1 2019-07-12
Release Candidate 1
## v1.0.0-snapshot-1 2019-07-09

### Added
- `SecretDefinitions` created via `CustomResourceDefinitions`
- If the `SecretDefinion` gets deleted, the corresponding secret will be removed too.
- New zap logger based on [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) project. Use `-enable-debug-log` to get a more verbose output.
### Fixes
- [#2 Switch to custom resource definitions instead of a single configmap](https://github.com/tuenti/secrets-manager/issues/2)
- [#8 Secrets deletion proposal](https://github.com/tuenti/secrets-manager/issues/8)

### Breaking changes
- congimaps won't be supported to define secrets, and so that won't work all the relevant configmap flags.
- log.format and log.level flags won't work anymore, as we have changed the logger to addapt to the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) project. Use `-enable-debug-log` to get a more verbose output.
- `config.backend-scrape-interval` no longer works as we check the backend state on every reconcile event. Use `reconcile-period` instead
- `listen-address` removed in favor of `metrics-addr`

## v1.0.0-snapshot 2019-05-22

### Added
- Enable Vault AppRole auth method and `secrets-manager` will try to re-login every time it fails to fetch the token. This will make `secrets-manager` more resilient to issues connecting to Vault that potentially caused the token to expire.
- New `secrets_manager_login_errors_total` Prometheus metric.

### Fixes
- [#27-Implement AppRole auth](https://github.com/tuenti/secrets-manager/issues/27)

### Breaking changes
- Token based login won't be supported, as re-login with and invalid token won't make `secrets-manager` to self-heal.
- This makes this new version not backward compatible with previous v0.2.0

## v0.2.0 - 2019-03-29

Stable
## v0.2.0-rc.2 - 2019-01-29

### Added
- New `secrets_manager_vault_max_token_ttl` metric, so a user could alert based on this and `secrets_manager_token_ttl`
- New `secrets_manager_secret_last_sync_status` metric, that shows wether the secret succeeded or not in last synchronization iteration

### Fixed
- Backend timeout not properly set through flags
- Deprecates `secrets_manager_vault_token_expired` metric as it was quite confusing since it's not really possible for `secrets-manager` to know when the token it's expired, just when it's "close to expire".
- Renames counter metrics to follow the Prometheus naming standard with the `_total` suffix instead of `_count`.
- Simplifies prometheus token renewal metrics by merging `secrets_manager_vault_token_lookup_errors_count` and `secrets_manager_vault_token_renew_errors_count` into one single metric `secrets_manager_vault_token_renewal_errors_total` with one more dimension called `vault_operation` which will be one of `lookup-self, renew-self, is-renewable`.

## v0.2.0-rc.1 - 2019-01-21

### Added
- Enable prometheus metrics
- `cfg.backend-timeout` flag to specify a connection timeout to the secrets backend.
- `listen-address` flag to specify the listen address of the HTTP API

### Fixed
- Bad return condition in startTokenRenewer, so token lookup won't
  happen in case of a token revoked.
