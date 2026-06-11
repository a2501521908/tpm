# TPM — Terminal Password Manager

> 中文文档： [README-zh.md](README-zh.md)

A lightweight, cross-platform CLI for **temporary / dynamic passwords** (TOTP,
tokens), designed for the AI era: terminal- and script-friendly, with clean
pipeline output that AI agents can capture directly.

TPM is **not** a long-term static password vault (like 1Password). It focuses on
the short-lived codes and OTPs that automation scripts and AI agents need while
running tasks.

## Highlights

- **Single static binary** — Go, no runtime, no dependencies. Drop it on `PATH`.
- **Cross-platform** — macOS, Windows, Linux (amd64 + arm64).
- **Encrypted at rest** — every entry is an independent `AES-256-GCM` `.enc`
  file. Nothing readable ever touches your cloud drive.
- **Cloud sync without lock-in** — point the data directory at iCloud / OneDrive /
  Google Drive; their official clients handle the syncing. One file per entry
  means no merge conflicts across devices.
- **Master key in the OS keyring** — macOS Keychain / Windows Credential Manager
  / Linux Secret Service. The key never lands on disk or in the cloud.
- **AI / pipeline friendly** — `--silent` prints only the code (no newline, no
  color, no decoration).

## Install

### Build from source

```bash
git clone https://github.com/zhangshuaike/tpm.git
cd tpm
make build        # produces ./bin/tpm
# or install onto your PATH:
make install
```

### Cross-compile release binaries

```bash
make release      # outputs dist/tpm-<os>-<arch>[.exe]
```

### Linux build prerequisite

`go-keyring` talks to the Secret Service over D-Bus. On Debian/Ubuntu install the
dev headers before building:

```bash
sudo apt-get install -y libsecret-1-dev
```

At runtime a Secret Service provider (GNOME Keyring, KWallet, etc.) must be
available.

## Quick start

### 1. Initialize

Point the data directory at your cloud sync folder. The master key is generated
and stored in the OS keyring automatically.

```bash
# macOS + iCloud
tpm init --dir "~/Library/Mobile Documents/com~apple~CloudDocs/tpm-data"

# Windows + OneDrive
tpm init --dir "%USERPROFILE%\OneDrive\tpm-data"
```

On a **second device**, transfer the master key securely and import it:

```bash
# On the primary device, print the key (handle with care):
#   (export helper is available via the keyring package; see Roadmap)
tpm init --dir "~/OneDrive/tpm-data" --import
# paste the base64 master key when prompted
```

### 2. Add a credential

Credentials are organized **by type**: `code` (dynamic TOTP) or `password`
(static). A single name can hold several types.

```bash
# A dynamic code (TOTP)
tpm add code github --secret "JBSWY3DPEHPK3PXP" --desc "GitHub 2FA"

# Give one name both a code and a password
tpm add code     tencent --secret "JBSWY3DPEHPK3PXP"
tpm add password tencent --secret "s3cr3t!"
```

### 3. Read a value

Grammar is `tpm <type> <name>`. When a name has only one type, the bare
`tpm <name>` shortcut works too.

```bash
# Explicit type
$ tpm code github
[TPM] Generating code for github...
Code: 482910 (Expires in 18s)

# Shortcut (only one type on this name)
$ tpm github
Code: 482910 (Expires in 18s)

# A name with several types — pick one
$ tpm code tencent
$ tpm password tencent

# AI / script mode — only the value, no newline
$ tpm code github --silent
482910

# Inject straight into an environment variable
export MY_TOKEN=$(tpm code github --silent)
```

### 4. Manage entries

```bash
tpm list                               # names and their types, e.g. "tencent [code, password]"
tpm show tencent                       # inspect an entry's types/metadata (no secrets)
tpm rename github vk             # rename a whole entry
tpm rename tencent tencent-pw --type password  # move just one type to another name
tpm rm tencent --type password         # remove just one type
tpm rm github                    # delete the whole entry
```

## How it works

```
<data_dir>/                  # inside your cloud sync folder
├── .tpm_meta                # non-sensitive metadata (version, crypto)
└── entries/
    ├── github-2fa.enc       # AES-256-GCM ciphertext (nonce-prefixed)
    └── prod-mysql.enc

~/.config/tpm/env.toml       # per-machine: points to <data_dir>
OS keyring (service=tpm)     # holds the 32-byte master key
```

Reading a value: load `data_dir` from `env.toml` → fetch the master key from the
keyring → read `entries/<name>.enc` → decrypt → pick the credential of the
requested type → run its provider → print. One `<name>.enc` file can hold
multiple credential types.

## Credential types

| Type       | Status   | Description                          |
| ---------- | -------- | ------------------------------------ |
| `code`     | ✅       | dynamic one-time code (RFC 6238 TOTP, 6 digits / SHA1 / 30s); compatible with Google & Microsoft Authenticator |
| `password` | ✅       | static password / token, returned verbatim |
| `script`   | planned  | run a custom command to produce a value |

New types plug in by implementing `provider.PasswordProvider` and calling
`provider.Register` — no changes to the core CLI required.

## Commands

| Command                              | Description                                 |
| ------------------------------------ | ------------------------------------------- |
| `tpm init --dir <path>`              | Initialize store + provision master key     |
| `tpm add <type> <name> --secret ...` | Add/update a credential of a type on a name |
| `tpm <type> <name> [--silent]`       | Print the value (e.g. `tpm code github`) |
| `tpm <name> [--silent]`              | Shortcut when the name has a single type    |
| `tpm list`                           | List names and their types                  |
| `tpm show <name>`                    | Inspect an entry's types/metadata (no secrets) |
| `tpm rename <old> <new> [--type <type>]` | Rename an entry, or move one type to another name |
| `tpm rm <name> [--type <type>]`      | Delete an entry, or just one type           |

## Development

```bash
make test    # run unit tests
make vet     # static analysis
make fmt     # format
```

## Roadmap

- `script` provider (run a custom command to produce a value)
- `tpm export-key` for safe secondary-device onboarding
- Optional local confirmation (Touch ID / prompt) before AI reads sensitive codes
- Pluggable storage backends (S3 / Git)

## License

MIT
