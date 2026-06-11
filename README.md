# space-permission-extractor

```
 _  _                         _   
| \| |_  _ _ __  ____ __  ___| |_ 
| .` | || | '  \(_-< '_ \/ _ \  _|
|_|\_|\_,_|_|_|_/__/ .__/\___/\__|
                   |_|            
    -- Permission Extractor --
```

TUI Tool to extract permission, role and ACL of all users in a Numspot Space.

## Features

- Interactive configuration with masked password input
- 4 extraction modes:
  - **Users**: Extract roles, permissions, and ACLs for all users
  - **Service Accounts**: Extract for all service accounts
  - **By Resource**: List identities with access to a specific resource
  - **By Identity**: List resources accessible by a specific user
- Real-time progress logs
- Scrollable results table with colored headers
- **Live search** with `/` in results view
- **CSV export** to `~/Downloads` folder

## Installation

### From Source

```bash
git clone https://github.com/numspot/space-permission-extractor.git
cd space-permission-extractor
make build
```

### From Releases

Download the latest release for your platform from the [Releases](https://github.com/numspot/space-permission-extractor/releases) page.

```bash
# Linux/macOS
tar -xf permission-extractor-tui_v1.x.x_linux_amd64.tar.xz

# Windows
# Extract the zip and run permission-extractor-tui.exe
```

## Usage

### Interactive Mode

```bash
./permission-extractor-tui
```

### With Pre-filled Options

```bash
./permission-extractor-tui --clientId <CLIENT_ID> --space <SPACE_ID>
```

### Using a `.env` File

Create a `.env` file in the application directory with the following variables:

```env
NUMSPOT_CLIENT_ID=your-oauth2-client-id
NUMSPOT_CLIENT_SECRET=your-oauth2-client-secret
NUMSPOT_SPACE_ID=your-space-uuid
# Use the URL of your target environment (preprod, prod...)
NUMSPOT_BASE_URL=https://api.eu-west-2.numspot.com
```

Then run:

```bash
./permission-extractor-tui
```

The app will automatically load the `.env` file. If all required fields are present, the configuration screen is skipped. Press `Ctrl+S` at any time to edit settings.

### Command Line Options

| Option | Description |
|--------|-------------|
| `--clientId <id>` | OAuth2 Client ID |
| `--clientSecret <secret>` | OAuth2 Client Secret |
| `--space <uuid>` | Space UUID |
| `--baseUrl <url>` | API Base URL (default: https://api.eu-west-2.numspot.com) |
| `--version, -v` | Show version information |
| `--help, -h` | Show help |

### Version

```bash
./permission-extractor-tui --version
# Output: permission-extractor-tui v1.0.0 (rev: abc1234, commit: ..., built: 2024-01-01, by: goreleaser)
```

## Keyboard Navigation

| Screen | Keys |
|--------|------|
| Config | Tab: next field, Shift+Tab: prev, Enter: continue, Esc: quit |
| Mode | Up/Down: select, Enter: start, Ctrl+S: settings, Esc: back |
| Progress | Esc: cancel |
| Results | Up/Down: scroll, Enter: export, `/`: search, Ctrl+S: settings, Esc: back |
| Export | Enter: save, Ctrl+S: settings, Esc: back |

## Required Permissions

Your OAuth2 client needs:
- `iam.user.get`: List users
- `iam.getPolicy`: Read IAM policies
- `iam.permission.get`: Read permission details
- `iam.role.get`: Read role details

## Output

Generates a CSV file with columns:
- User Mode: Entity UUID, Entity Name, Auth Type, Auth ID, Auth Name, Description, Action
- Resource Mode: Identity UUID, Identity Type, Domain, Resource Type, Auth Type, Auth ID, Auth Name, Description, Action
- Identity Mode: Resource UUID, Domain, Resource Type, Auth Type, Auth ID, Auth Name, Description, Action

## Development

```bash
make build    # Build binary
make run      # Build and run
make test     # Run tests
make lint     # Run linter
make clean    # Clean build artifacts
make release-snapshot  # Test release locally (no publish)
```

## Release

Releases are automated via [GoReleaser](https://goreleaser.com/) and [GitHub Actions](https://github.com/features/actions).

To create a new release:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The CI pipeline will automatically build cross-platform binaries and publish them to the [Releases](https://github.com/numspot/space-permission-extractor/releases) page.

## License

Apache-2.0 License © NumSpot
