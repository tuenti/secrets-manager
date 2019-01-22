## v0.2.0-rc.1 - 2019-01-21

### Added
- Enable prometheus metrics
- `cfg.backend-timeout` flag to specify a connection timeout to the secrets backend.

### Fixed
- Bad return condition in startTokenRenewer, so token lookup won't
  happen in case of a token revoked.
