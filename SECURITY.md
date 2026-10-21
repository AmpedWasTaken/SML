# Security policy

SML parses markup and assets that may originate from third parties. Resource limits and the boundary between declarative action metadata and host-side execution are security-critical.

## Supported versions

The project is pre-1.0. Security fixes target the current `main` branch and any explicitly maintained release series; there is no promise of fixes to older commits.

## Reporting a vulnerability

**Do not publish proof-of-concept exploits, sensitive data, or vulnerability details in a public issue.**

Prefer [GitHub private vulnerability reporting](https://github.com/AmpedWasTaken/SML/security/advisories/new) when that option is enabled. Otherwise, contact the repository maintainer privately through a contact channel they have published on their GitHub profile. If neither channel is available, request a private reporting channel without revealing the vulnerability.

Include the affected commit/version, a minimal reproduction, expected and actual behavior, and impact. Avoid real credentials or harmful tests. Maintainers will assess reports and coordinate disclosure when a fix is available.

## Threat boundaries

SML does not execute shell commands, fetch remote content, dispatch plugin actions, or authorize user access. Hosts must authorize every action independently of SML metadata, and treat bindings and local assets according to their application's threat model.

Read [docs/security.md](docs/security.md) for integration guidance.
