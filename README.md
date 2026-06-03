# NumSpot Permission Extractor TUI

Terminal User Interface for extracting permissions from NumSpot spaces.

## Features

- Interactive configuration with masked password input
- 4 extraction modes:
  - **Users**: Extract roles, permissions, and ACLs for all users
  - **Service Accounts**: Extract for all service accounts
  - **By Resource**: List identities with access to a specific resource
  - **By Identity**: List resources accessible by a specific user
- Real-time progress logs
- Scrollable results table
- CSV export

## Installation

```bash
git clone <repo-url>
cd permission-extractor-tui
make build
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

### Command Line Options

| Option | Description |
|--------|-------------|
| `--clientId <id>` | OAuth2 Client ID |
| `--space <uuid>` | Space UUID |
| `--baseUrl <url>` | API Base URL (default: https://api.eu-west-2.numspot.com) |
| `--help, -h` | Show help |

## Keyboard Navigation

| Screen | Keys |
|--------|------|
| Config | Tab: next field, Shift+Tab: prev, Enter: continue, q: quit |
| Mode | Arrow keys: select, Enter: start, q: back |
| Progress | q: cancel |
| Results | Arrows: scroll, Enter: export, q: back |
| Export | Enter: save & quit, q: back |

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

## License

NumSpot Proprietary License
