# Shadowbox Security Overview (Beginner's Guide)

This document explains the security improvements in Shadowbox in plain language.
It is meant as a reference you can read at your own pace — no security background
required.

For operational steps (signing setup, releases, branch protection), see
[RELEASING.md](RELEASING.md). For reporting vulnerabilities, see
[SECURITY.md](../SECURITY.md).

---

## The big question

Whenever you download software or accept code from GitHub, you are really asking:

> **Did this really come from the person I trust, and was it changed by anyone else?**

Security hardening adds checks at every step so you — and your users — can answer
"yes" with evidence, not just hope.

---



## Part 1: Protecting the app itself (runtime)

These are fixes inside Shadowbox so the program cannot be tricked into doing
harmful things on your machine.

### Secrets in config files

Shadowbox stores optional API keys (Genius) in:

`~/.config/shadowbox/config.yaml`

**Before:** the file could be readable by other users on your computer.

**Now:** the config directory is `0700` and the file is `0600` — only your user
account can read it. Think of it like locking a drawer that holds your passwords.

### yt-dlp argument injection

Shadowbox runs an external program called `yt-dlp` with your search query or URL.

**The risk:** a malicious input starting with `-` (for example `--exec ...`) could
be interpreted as a command flag instead of plain text. Some flags can run
arbitrary commands.

**The fix:**

- Inputs are validated before use (no control characters, allowed hosts only).
- A `--` separator is inserted before your query/URL so `yt-dlp` treats it as
data, never as a flag.



### TLS / HTTPS for downloads

**Before:** some download strategies disabled certificate checking
(`--no-check-certificates`), which makes "man-in-the-middle" attacks easier.

**Now:** certificates are always verified — like checking ID before accepting a
package delivery.

### Cover art downloads

**Before:** Shadowbox could download unlimited data from any URL returned by a
metadata API.

**Now:**

- URL must be `https://`
- HTTP response must be status 200
- Content-Type must be an image
- Download is capped at 10 MiB

This limits abuse if a bad URL ever sneaks into the metadata pipeline.

---



## Part 2: Protecting the GitHub repository (source code)

This is about stopping bad or tampered code from landing in your project.

### Signed commits (SSH signing)

Normally Git records who committed (`Author: Julian`), but that name is just text —
anyone can set it.

**Signing** cryptographically proves *you* made the commit, using your SSH private
key. GitHub shows a green **Verified** badge on signed commits.

See [SSH keys: authentication vs signing](#ssh-keys-authentication-vs-signing) below.

### Branch protection

Rules on the `main` branch, such as:

- No direct pushes — changes must go through a pull request
- CI must pass before merge
- Commits must be signed
- No force-push or branch deletion

Think of it as requiring review and checks before changing the master copy.

### CODEOWNERS

The file `.github/CODEOWNERS` tells GitHub that `@EnJulian` owns the repo. GitHub
will automatically request your review on pull requests.

### Dependabot

A bot that opens pull requests when Go dependencies or GitHub Actions have updates.
This helps you patch known vulnerabilities instead of running old, risky versions.

### CI security scans


| Tool            | What it does                                                            |
| --------------- | ----------------------------------------------------------------------- |
| **govulncheck** | Scans Go code for known CVEs in dependencies                            |
| **gitleaks**    | Scans git history for accidentally committed secrets (API keys, tokens) |
| **CodeQL**      | GitHub's static analysis — looks for security bugs in your code         |


These run automatically on pushes and pull requests to `main`.

### Pinned GitHub Actions

Workflows used to reference actions like `@main` or `@v4`. Those tags can change
silently — someone could swap in malicious code.

**Now:** every action is pinned to an exact commit SHA, so the same code runs
every time. Dependabot can still propose updates via pull requests.

---



## Part 3: Protecting releases (what users download)

When someone runs `brew install shadowbox` or downloads from GitHub Releases,
they want proof the binary was not tampered with.

### Checksums (`checksums.txt`)

A file listing SHA-256 hashes of every release archive. If one byte of a binary
changes, the hash will not match — tampering becomes detectable.

### Cosign signing (keyless)

GoReleaser signs `checksums.txt` with **cosign** during the GitHub Actions
release workflow.

**Keyless** means there is no private key file sitting on a disk somewhere. The
signature is tied to your GitHub Actions workflow via OIDC — GitHub proves the
signer is your release pipeline.

Users verify with:

```bash
cosign verify-blob checksums.txt \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/EnJulian/shadowbox/.github/workflows/release.yml@refs/tags/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

Think of it as a notarized stamp on the checksum file.

### SBOM (Software Bill of Materials)

A manifest listing every dependency bundled into each release archive. If a
vulnerability is discovered in a library later, you can check whether a given
release was affected.

### GitHub attestations (SLSA provenance)

GitHub records *how* each binary was built: which workflow, which commit, which
runner. Verifiable with:

```bash
gh attestation verify shadowbox-linux-amd64.tar.gz \
  --owner EnJulian --repo shadowbox
```

This proves the binary was built by your CI from your source — not compiled on
some unknown machine.

### Signed release tags

`git tag -s v1.3.0` creates a release tag signed with your SSH key. This proves
the tag (which triggers the automated release) came from you.

---



## How it all fits together

```
You write code
    → sign commits with SSH key
    → open pull request
        → CI runs tests, lint, govulncheck, gitleaks, CodeQL
    → merge to main (branch protection enforced)
    → create signed tag vX.Y.Z
        → GoReleaser builds binaries
        → syft generates SBOMs
        → cosign signs checksums
        → GitHub publishes provenance attestations
    → user downloads release
        → verifies cosign signature + attestation
        → installs shadowbox
```

---



## Quick reference table


| Thing              | One-sentence explanation                             |
| ------------------ | ---------------------------------------------------- |
| SSH commit signing | Proves you wrote the commit                          |
| Branch protection  | Stops bad code landing on `main` without review      |
| gitleaks           | Catches accidentally committed passwords             |
| govulncheck        | Finds known bugs in your dependencies                |
| CodeQL             | Finds security bugs in your code                     |
| Pinned actions     | Stops CI from running surprise changed code          |
| Cosign             | Proves release checksums were not tampered with      |
| SBOM               | Lists everything inside each binary                  |
| Attestations       | Proves binaries were built by your CI from your repo |
| Signed tags        | Proves the release version tag came from you         |


---



## SSH keys: authentication vs signing

This is the most common point of confusion on GitHub.

### Two different jobs


| Type                   | Purpose                                            | How you use it                          |
| ---------------------- | -------------------------------------------------- | --------------------------------------- |
| **Authentication key** | Log in to GitHub over SSH (`git push`, `git pull`) | `git@github.com:EnJulian/shadowbox.git` |
| **Signing key**        | Prove a commit or tag was made by you              | `git commit -S`, `git tag -s`           |


They can use the **same key file** on your machine (`~/.ssh/id_ed25519`), but
GitHub treats them as **separate registrations**.

### How to tell what you already have

Go to **GitHub → Settings → SSH and GPG keys**.

GitHub shows two sections:

1. **Authentication keys** — usually labeled with access like **Read/write** or
  **Read-only**. These are for push/pull.
2. **Signing keys** — used only for the green **Verified** badge on commits and
  tags.

Your existing key showing **"Never used — Read/write"** is an **authentication
key**. It is **not** registered as a signing key yet. That is normal if you set up
SSH for `git push` but never configured commit signing.

### What to do (reuse your existing key)

You do **not** need to generate a new key or overwrite `~/.ssh/id_ed25519`.

**Step 1 — On your machine**, configure Git to sign with the existing key:

```bash
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
git config --global tag.gpgsign true

mkdir -p ~/.config/git
echo "$(git config user.email) namespaces=\"git\",$(cat ~/.ssh/id_ed25519.pub)" \
  >> ~/.config/git/allowed_signers
git config --global gpg.ssh.allowedSignersFile ~/.config/git/allowed_signers
```

**Step 2 — On GitHub**, register the same public key as a signing key:

1. Copy your public key: `cat ~/.ssh/id_ed25519.pub`
2. Go to **Settings → SSH and GPG keys**
3. Click **New SSH key** (GitHub's label — signing keys are added here too)
4. Set **Key type** to **Signing Key** (not Authentication Key)
5. Paste the public key and save

GitHub allows the same public key to exist as both an authentication key and a
signing key. You upload it twice with different key types.

**Step 3 — Verify**:

```bash
git commit --allow-empty -S -m "test signed commit"
git log -1 --show-signature
git push   # authentication key handles this
```

On GitHub, the commit should show **Verified**.

### Why GitHub only shows "Add SSH key" and "Add GPG key"

- **Add SSH key** — covers both authentication and signing SSH keys. When adding,
you choose the key type in the form (Authentication Key vs Signing Key).
- **Add GPG key** — the older signing method. You chose SSH signing instead, so
you can ignore GPG unless you prefer it.

---



## What you still need to do manually

These require your GitHub account and cannot be done from code alone:

1. **Register your SSH public key as a Signing key** (see above)
2. **Configure Git signing locally** (see above)
3. **Apply branch protection** — `gh` commands in [RELEASING.md](RELEASING.md#github-repository-security)
4. **Enable secret scanning** — repo Settings → Code security and analysis

---



## Further reading

- [RELEASING.md](RELEASING.md) — release process, verification commands, branch protection
- [SECURITY.md](../SECURITY.md) — vulnerability disclosure policy
- [Sigstore cosign docs](https://docs.sigstore.dev/cosign/overview/)
- [GitHub SSH commit verification](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification)

