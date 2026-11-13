# multi-git — multiple Git identities on one machine

**Short description:** Configure multiple Git identities (personal and work) on one machine using conditional includes, SSH keys, and signed commits — with a cross-platform CLI.

**Longer description:** `multi-git` helps developers safely use different Git identities (name, email, SSH key, and commit signing) on a single machine. It provides documented templates for folder- and URL-based conditional Git configuration, SSH setup, and an optional cross-platform Go CLI that detects existing tools, generates configuration files, creates backups, and verifies the result. Designed for Windows, WSL, Linux, and macOS.

> **Status:** Phase 1 — configuration formalized and documented.  
> Phase 2 (automation CLI) is in progress under `./cli`.

## Problem

On a single machine there are personal and corporate repositories. Without explicit
configuration, Git may reuse a global identity and create commits (including signed
ones) attributed to the wrong user.

## Solution

Use Git's **`includeIf`** feature to override `user.*` and `commit.gpgsign` based on
the repository folder, and map each SSH host to the correct key.

Example identities used throughout this guide (replace with your own):

| Account     | Name          | Email             | Folder       | SSH key      | Primary remote host        |
|-------------|---------------|-------------------|--------------|--------------|----------------------------|
| Personal    | Your Name     | you@example.com   | `~/personal` | `id_ed25519` | `github.com`               |
| Corporate   | Your Name     | you@company.com   | `~/work`     | `id_rsa`     | Azure DevOps (`*.visualstudio.com`, `*.azure.com`) |

## Prerequisites

- Git **2.13+** (for `includeIf`) and **2.34+** for SSH commit signing.
- OpenSSH client with `ssh-agent`.
- SSH keys already generated (never commit private keys).

Check:

```bash
git --version
ssh -V
```

## Folder layout

```text
~/personal/   # personal repositories (GitHub)
~/work/       # corporate repositories (Azure DevOps)
```

Create them if needed:

```bash
mkdir -p ~/personal ~/work
```

Windows (PowerShell):

```powershell
New-Item -ItemType Directory -Force -Path "$HOME\personal", "$HOME\work"
```

## Placeholders

Templates use `{{...}}` placeholders. Replace before copying:

| Placeholder              | Example value     |
|--------------------------|-------------------|
| `{{PERSONAL_NAME}}`      | `Your Name`       |
| `{{PERSONAL_EMAIL}}`     | `you@example.com` |
| `{{PERSONAL_SSH_KEY}}`   | `id_ed25519`      |
| `{{CORPORATE_NAME}}`     | `Your Name`       |
| `{{CORPORATE_EMAIL}}`    | `you@company.com` |
| `{{CORPORATE_SSH_KEY}}`  | `id_rsa`          |

Store your real values in `private/identity.env` (git-ignored). See `.gitignore`.

## Installation — Git configuration

1. Copy the templates:

   ```bash
   cp gitconfig/.gitconfig ~/.gitconfig
   cp gitconfig/.gitconfig-personal ~/.gitconfig-personal
   cp gitconfig/.gitconfig-work ~/.gitconfig-work
   ```

   On Windows (Git Bash), use the same commands with POSIX paths (`~`).

2. Replace placeholders in all three `~/.gitconfig*` files.

3. (Optional) Adjust `core.autocrlf` in `~/.gitconfig`:
   - Windows: `true`
   - Linux/macOS/WSL: `input`

## Installation — SSH

1. Copy the template:

   ```bash
   mkdir -p ~/.ssh
   cp ssh/config ~/.ssh/config
   chmod 600 ~/.ssh/config
   ```

2. Replace placeholders in `~/.ssh/config`.

3. Create the allowed signers file (for verifying signed commits):

   ```bash
   : > ~/.ssh/allowed_signers
   printf '%s %s\n' "you@example.com" "$(cat ~/.ssh/id_ed25519.pub)" >> ~/.ssh/allowed_signers
   printf '%s %s\n' "you@company.com" "$(cat ~/.ssh/id_rsa.pub)" >> ~/.ssh/allowed_signers
   chmod 600 ~/.ssh/allowed_signers
   ```

4. Ensure key permissions:

   ```bash
   chmod 600 ~/.ssh/id_ed25519 ~/.ssh/id_rsa
   chmod 644 ~/.ssh/id_ed25519.pub ~/.ssh/id_rsa.pub
   ```

## How it works

- `~/.gitconfig` defines the default identity (personal) and the `includeIf` rules.
- Repositories under `~/personal/` load `~/.gitconfig-personal`.
- Repositories under `~/work/` load `~/.gitconfig-work`.
- Each override sets `user.name`, `user.email`, `user.signingkey`, and `commit.gpgsign = true`.
- SSH uses `IdentitiesOnly yes` to avoid offering the wrong key.

## Verification

Inside a personal repository:

```bash
git config --show-origin user.name
git config --show-origin user.email
git config --show-origin user.signingkey
git config --show-origin commit.gpgsign
```

Expected: values coming from `~/.gitconfig-personal`.

Inside a corporate repository: values coming from `~/.gitconfig-work`.

SSH connectivity:

```bash
ssh -T git@github.com
ssh -T git@vs-ssh.visualstudio.com
```

Signature check (after a commit):

```bash
git verify-commit HEAD
# or
git log --show-signature -1
```

## Alternative: separation by remote URL (`hasconfig`)

If you do not want to separate by folder, use rules based on the remote URL
(uncomment the matching block in `~/.gitconfig`):

```ini
[includeIf "hasconfig:remote.*.url:https://github.com/**"]
    path = ~/.gitconfig-personal

[includeIf "hasconfig:remote.*.url:https://dev.azure.com/**"]
    path = ~/.gitconfig-work

[includeIf "hasconfig:remote.*.url:git@ssh.dev.azure.com:**"]
    path = ~/.gitconfig-work
```

Adjust the patterns to match `git remote -v`.

## Troubleshooting

- **Wrong identity in a repository:** confirm the repository is under `~/personal/` or
  `~/work/` (real path, no unexpected symlink) and run `git config --show-origin`.
- **Cloned with `gh` (GitHub CLI):** `gh repo clone` may write `user.name` and
  `user.email` (and sometimes `user.signingkey` / `commit.gpgsign`) directly into the
  repository's `.git/config`. Local config has higher precedence than the global
  `includeIf` rules, so the folder-based override is silently ignored. Remove the
  local overrides so `includeIf` can apply:
  ```bash
  git config --local --unset user.name
  git config --local --unset user.email
  git config --local --unset user.signingkey
  git config --local --unset commit.gpgsign
  ```
  Then check with `git config --show-origin user.email`.
- **`includeIf` does not apply on Windows:** use forward slashes `/` and the `~/`
  prefix; avoid `C:\...` in `gitdir`. Prefer `gitdir/i:` (case-insensitive).
- **Commit not signed:** check `commit.gpgsign = true` in the override and that the
  public key is registered as a signing key on GitHub/Azure DevOps.
- **`git verify-commit` fails:** check `~/.ssh/allowed_signers` and file permissions.
- **SSH offers the wrong key:** confirm `IdentitiesOnly yes` and `IdentityFile`.
- **Repository outside the default folders:** move it or use the URL-based alternative.

## Security

- Never commit private keys or filled-in `~/.gitconfig-*` files.
- This repository contains only templates with placeholders.
- Rotate keys immediately if they are leaked.

## Cross-platform notes

- **Linux/macOS/WSL:** POSIX paths (`~/.gitconfig`, `~/.ssh/config`).
- **Windows:** Git for Windows accepts `~` and `/`; use Git Bash for copy commands.
  PowerShell can be used to create folders.
- **WSL:** treat as Linux; repositories accessed via `/mnt/...` may not match
  `~/...` — prefer cloning inside the WSL filesystem.

## Phase 2 (in progress)

The cross-platform CLI lives in `./cli` (Go). It:
- detects SSH keys, `gh`, `az`, and existing configuration;
- generates files from embedded templates;
- backs up and asks for confirmation before overwriting;
- validates the installation (`verify`).

Run without installing:
- Linux/macOS/WSL: `curl -fsSL .../cli/scripts/install.sh | bash`
- Windows: `irm .../cli/scripts/install.ps1 | iex`

See `./cli/README.md`.