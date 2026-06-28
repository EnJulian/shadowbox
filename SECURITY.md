# Security Policy

## Supported versions

| Version | Supported |
| ------- | --------- |
| Latest release | Yes |
| Previous major | Best effort |
| Older | No |

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Report issues privately through [GitHub Security Advisories](https://github.com/EnJulian/shadowbox/security/advisories/new). Include:

- A description of the issue and its impact
- Steps to reproduce
- Affected versions
- Any proof-of-concept or suggested fix (optional)

## Response expectations

- **Acknowledgement:** within 5 business days
- **Initial assessment:** within 10 business days
- **Fix or mitigation plan:** communicated as soon as a timeline is known

We will coordinate disclosure with you and credit reporters who wish to be named when a fix is released.

## Supply-chain verification

Release artifacts are signed with Sigstore cosign (keyless) and include GitHub build provenance attestations. See [docs/RELEASING.md](docs/RELEASING.md#verifying-downloads) for verification steps.
