# space-permission-extractor

```
 _  _                         _   
| \| |_  _ _ __  ____ __  ___| |_ 
| .` | || | '  \(_-< '_ \/ _ \  _|
|_|\_|\_,_|_|_|_/__/ .__/\___/\__|
                   |_|            
    -- Permission Extractor --
```

Interactive TUI for auditing IAM permissions, roles and ACLs across all identities (users & service accounts) within a NumSpot space. Exports results to CSV for compliance and security reviews.

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

### From Releases (Recommended)

1. Download the latest release for your platform from [GitHub Releases](https://github.com/numspot/space-permission-extractor/releases).

2. Extract the archive:
```bash
# Linux/macOS
tar -xf permission-extractor-tui_v1.x.x_linux_amd64.tar.xz
cd permission-extractor-tui/

# Windows
# Extract the .zip and open the folder
```

3. Create a `.env` file in the same folder as the binary:
```bash
cat > .env << EOF
NUMSPOT_CLIENT_ID=your-oauth2-client-id
NUMSPOT_CLIENT_SECRET=your-oauth2-client-secret
NUMSPOT_SPACE_ID=your-space-uuid
EOF
```

4. Run the tool:
```bash
./permission-extractor-tui
```

### From Source

```bash
git clone https://github.com/numspot/space-permission-extractor.git
cd space-permission-extractor
make build
```

Then create a `.env` file in the project directory and run `./permission-extractor-tui`.

## Usage

### Interactive Mode

```bash
./permission-extractor-tui
```

### With Pre-filled Options

```bash
./permission-extractor-tui --clientId <CLIENT_ID> --space <SPACE_ID>
```

### `.env` File Variables

If you need to customize the configuration, edit the `.env` file with these variables:

```env
NUMSPOT_CLIENT_ID=your-oauth2-client-id
NUMSPOT_CLIENT_SECRET=your-oauth2-client-secret
NUMSPOT_SPACE_ID=your-space-uuid
# Use the URL of your target environment (preprod, prod...)
NUMSPOT_BASE_URL=https://api.eu-west-2.numspot.com
```

The app will automatically load the `.env` file from the current directory. Fields are pre-filled from the `.env` file or command-line flags. Press `Enter` to continue to the mode selection, or `Ctrl+S` at any time to edit settings.

### Command Line Options

| Option | Description |
|--------|-------------|
| `--clientId <id>` | OAuth2 Client ID |
| `--clientSecret <secret>` | OAuth2 Client Secret |
| `--space <uuid>` | Space UUID |
| `--baseUrl <url>` | API Base URL (default: https://api.eu-west-2.numspot.com) |
| `--version, -v` | Show version information |
| `--help, -h` | Show help |

**Note:** Extraction modes (Users, Service Accounts, By Resource, By Identity) are selected interactively in the TUI after launch. They are not CLI flags.

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
