# tins

**Temporary Instances** - A CLI tool for managing temporary OpenStack instances. This tool creates ephemeral instances with automatic SSH key management.

## Features

- **Ephemeral Instances**: Creates temporary instances with a consistent naming prefix (`tins-`) and Docker-style two-word names (e.g., `tins-mystical-honda`)
- **Automatic SSH Key Management**: Generates SSH keys automatically and stores them in `~/.ssh/`
- **Easy Cleanup**: Terminating an instance also removes its associated SSH keys
- **OpenStack Integration**: Uses the Gophercloud Go SDK for OpenStack API integration

## Installation

```bash
go build -o tins
```

## Configuration

The tool loads configuration from a YAML file and environment variables. Environment variables override values from the config file.

### Configuration File

Create a YAML configuration file in one of these locations (searched in order):
1. `.config/tins.yaml` (in the current directory where you run the tool)
2. `~/.config/tins/tins.yaml` (in your home directory)

Example `tins.yaml`:

```yaml
auth_url: "https://your.openstack.com/keystone/v3"
username: "username@domain.com"
domain_name: "default"

project_id: "project-id"
project_name: "project-name"

region_name: "region-name"
availability_zone: "availability-zone"

image_name: "image-name"
flavor_name: "m1.small"

network_name: "network-name"
network_attachment_mode: "existing_network"
```

### Required Environment Variables

- `OS_PASSWORD` - OpenStack password (must be set as environment variable, not in config file)

## Usage

### Check Version

```bash
tins version
```

Displays the version and build information.

### Create a Temporary Instance

```bash
tins create [instance-name]
```

If `instance-name` is not provided, a random Docker-style two-word name will be generated (e.g., `mystical-honda`).

This will:
1. Generate an SSH key pair and save it to `~/.ssh/tins-<instance-name>` and `~/.ssh/tins-<instance-name>.pub`
2. Create an OpenStack instance named `tins-<instance-name>` (e.g., `tins-mystical-honda`)
3. Tag the instance with `tins` metadata
4. Install the public key on the instance
5. Wait for the instance to become active
6. Display connection information

### List Temporary Instances

```bash
tins list
```

Lists all instances with `tins-` prefix or `tins: true` metadata.

### Connect to a Temporary Instance

```bash
tins connect
```

This will show an interactive menu of all available instances. You can:
- Use arrow keys to navigate
- Press Enter to select
- Press Esc or Ctrl+C to cancel

You can also connect directly by providing the instance name or ID:

```bash
tins connect mystical-honda
# or
tins connect tins-mystical-honda
# or
tins connect <instance-id>
```

The command automatically:
1. Finds the instance IP address
2. Uses the correct SSH key from `~/.ssh/tins-<instance-name>`
3. Connects as `root` user

### Terminate a Temporary Instance

```bash
tins terminate <instance-name-or-id>
# Examples:
tins terminate mystical-honda
# or
tins terminate tins-mystical-honda
# or
tins terminate <instance-id>
```

This will:
1. Terminate the OpenStack instance
2. Delete the associated SSH key pair from `~/.ssh/`

## Example Configuration

### Using Config File (Recommended)

1. Create the config directory:
```bash
mkdir -p ~/.config/tins
```

2. Copy the example config:
```bash
cp .config/tins.example.yaml ~/.config/tins/tins.yaml
```

3. Edit the config file with your values

4. Set only the password as an environment variable:
```bash
export OS_PASSWORD="your-password"
```

### Using Environment Variables Only

You can also set all values via environment variables (environment variables override config file):

```bash
export OS_AUTH_URL="https://your.openstack.com/keystone/v3"
export OS_PROJECT_ID="project-id"
export OS_REGION_NAME="region-name"
export OS_IMAGE_NAME="image-name"
export OS_NETWORK_NAME="network-name"
export OS_NETWORK_ATTACHMENT_MODE="existing_network"
export OS_AVAILABILITY_ZONE="availability-zone"
export OS_PROJECT_NAME="project-name"
export OS_USERNAME="your-username"
export OS_PASSWORD="your-password"
```

## SSH Key Management

SSH keys are automatically generated and stored in `~/.ssh/` with the naming convention:
- Private key: `~/.ssh/tins-<instance-name>`
- Public key: `~/.ssh/tins-<instance-name>.pub`

When you terminate an instance, both keys are automatically deleted.

## Instance Naming and Tagging

All temporary instances are created with:
- **Name prefix**: `tins-<instance-name>` (e.g., `tins-mystical-honda`)
- **Metadata**: `tins: true`
- **Auto-generated names**: If no name is provided, a Docker-style two-word name (adjective-noun) is generated

This makes it easy to identify and manage temporary instances in your OpenStack environment.

## Development

### Building from Source

```bash
go build -o tins .
```

### Running Tests

```bash
go test ./...
```

### CI/CD

This project uses GitHub Actions for CI/CD:

- **CI Workflow** (`ci.yml`): Runs on every push and pull request
  - Runs tests with race detection
  - Runs linting with golangci-lint
  - Builds binaries for multiple platforms

- **Release Workflow** (`release.yml`): Triggers on version tags (e.g., `v1.0.0`)
  - Builds binaries for Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64)
  - Creates a GitHub release with all binaries and checksums
  - Automatically detects pre-release versions (tags containing `-`)

### Creating a Release

To create a new release:

1. Create and push a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The release workflow will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Attach all binaries and checksums
