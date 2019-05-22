## v1.0.0-snapshot 2019-05-22

### Added
- Enable Vault AppRole auth method and `secrets-manager` will try to re-login every time it fails to fetch the token. This will make `secrets-manager` more resilient to issues connecting to Vault that potentially caused the token to expire.
- New `secrets_manager_login_errors_total` Prometheus metric.

### Breaking changes
- Token based login won't be supported, as re-login with and invalid token won't make `secrets-manager` to self-heal.
- This makes this new version not backward compatible with previous v0.2.0

## v0.2.0 - 2019-03-29

Stable
## v0.2.0-rc.2 - 2019-01-29

### Added
- New `secrets_manager_vault_max_token_ttl` metric, so a user could alert based on this and `secrets_manager_token_ttl`
- New `secrets_manager_secret_last_sync_status` metric, that shows wether the secret succeded or not in last synchronization iteration

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
