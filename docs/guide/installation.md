---
order: 1
---

# Installation

Stencil is written in [Go](https://golang.org/) with support for multiple platforms. The latest release can be found at [stencil Releases](https://github.com/rgst-io/stencil/releases).

Stencil releases are provided for most major platforms and architecture
combinations.

Stencil may also be compiled from source wherever the Go toolchain can run; e.g., on other operating systems such as DragonFly BSD, OpenBSD, Plan&nbsp;9, Solaris, and others. See <https://golang.org/doc/install/source> for the full set of supported combinations of target operating systems and compilation architectures.

## Quick Install

### Homebrew (macOS)

Stencil is installable via the [official Homebrew formula](https://formulae.brew.sh/formula/stencil#default)

```bash
brew install stencil
```

### APT (Linux [Debian & Ubuntu])

```bash
# Ensure that we're ready to install APT keyrings and that we have
# some pre-req binaries for setting up custom repos.
sudo apt update -y
sudo apt install -y gpg sudo wget curl
sudo install -dm 755 /etc/apt/keyrings

# Trust our signing key
wget -qO - https://pkg.rgst.io/apt/gpg.key | gpg --dearmor | \
  sudo tee /etc/apt/keyrings/stencil-archive-keyring.gpg 1>/dev/null

# Register our package repository
echo "deb [signed-by=/etc/apt/keyrings/stencil-archive-keyring.gpg] https://pkg.rgst.io/apt /" | \
  sudo tee /etc/apt/sources.list.d/stencil.list

# Update package list and install 'stencil'
sudo apt update
sudo apt install stencil
```

### Alpine (Linux [Alpine Linux])

```bash
apk add --no-cache sudo curl
echo "https://alpine.fury.io/rgst-io/" >> /etc/apk/repositories
curl https://alpine.fury.io/rgst-io/rgst-io@fury.io-946e9786.rsa.pub | sudo tee /etc/apk/keys/'rgst-io@fury.io-946e9786.rsa.pub' >/dev/null
```

### Github Action

Quickstart:

```yaml
- name: Install Stencil
  uses: rgst-io/stencil-action@latest
  with:
    github-token: ${{ github.token }}
    version: "latest"
```

[Github Action Documentation](https://github.com/marketplace/actions/stencil-action)

### Binary (Cross-platform)

Download the appropriate version for your platform from [Stencil Releases](https://github.com/rgst-io/stencil/releases). Once downloaded, the binary can be run from anywhere. You don't need to install it into a global location. This works well for shared hosts and other systems where you don't have a privileged account.

Ideally, you should install it somewhere in your `PATH` for easy use. `/usr/local/bin` is the most probable location.

### Source

Stencil is quite easy to build from source as well, should you not wish
to use pre-builts.

#### Prerequisite Tools

- Git
- [mise](https://mise.jdx.dev/getting-started.html#quickstart)

#### Fetch from GitHub

```bash
git clone https://github.com/rgst-io/stencil/stencil.git
cd stencil
mise install
mise run build
cp ./bin/stencil "$(go env GOPATH)/bin/stencil"
```

## Verifying Download Releases

Traditionally, one would provide a PGP key for verification. However, we
go a step beyond that.

Using Github attestations, you can verify a downloaded archive (and
thus, the binary) at any time. This proves that the binary originated
from Github Actions and that it was not tampered with at the time. This
requires the `gh` CLI to function, however.

We provide this for all of our top-level archive types (but not the
binary themselves, currently).

Here's an example:

```bash
# Downloaded the latest darwin_arm64 stencil release at the time.
$ ls -alh stencil_0.8.0_darwin_arm64.tar.gz
.rw-r--r-- jaredallard staff 6.7 MB Mon Jul 29 22:00:27 2024  stencil_0.8.0_darwin_arm64.tar.gz

$ gh attestation verify stencil_0.8.0_darwin_arm64.tar.gz --repo rgst-io/stencil
Loaded digest sha256:c6c9b63c0dff239182343b6bd0b3225597ee2bc393a358156a2bd8070090c33d for file://stencil_0.8.0_darwin_arm64.tar.gz
Loaded 1 attestation from GitHub API
✓ Verification succeeded!

sha256:c6c9b63c0dff239182343b6bd0b3225597ee2bc393a358156a2bd8070090c33d was attested by:
REPO             PREDICATE_TYPE                  WORKFLOW
rgst-io/stencil  https://slsa.dev/provenance/v1  .github/workflows/release.yaml@refs/heads/main
```

## Upgrading Stencil

Upgrading Stencil is as easy as downloading and replacing the executable
you’ve placed in your `PATH`, or updating with your favorite package
manager that you used to install it.
