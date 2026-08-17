## v0.3.6 (2026-08-17)

### Bug Fixes

* **ci:** use --version flag, not version subcommand (semrel has no version subcommand)
* remove commit_changelog:false with no replacement plugin - CHANGELOG.md was never updated
* resolve OCI digests via jq instead of oras tag lookup (Docker rewrites bare tags to :latest)
* extract OCI tar before oras resolve to fix ORAS 1.3.1 path parsing

### Other Changes

* harden Teams hook delivery (#13)

# Changelog