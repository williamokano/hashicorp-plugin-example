# HashiCorp Plugin Architecture Example

> **🤖 AI-Generated Project Notice**
> 
> This entire project is an **AI test case** to explore how well software can be written through "vibecoding" - providing high-level descriptions and letting AI handle the implementation. This codebase was created with **zero developer oversight** - every line of code, architecture decision, and documentation was generated exclusively by AI (Claude) based on conversational prompts.
> 
> The goal is to see how sophisticated and production-ready a project can become when developed entirely through AI-assisted "vibe-based" programming, where the human provides the vision and the AI handles all implementation details.

---

## About This Project

A demonstration of building a plugin-based CLI tool using HashiCorp's go-plugin library, similar to how Terraform manages its providers. Despite being entirely AI-generated, it implements a complete, production-grade plugin architecture.

## Features

- **Plugin Discovery**: Automatically discovers plugins with `plugin-` prefix in predefined paths
- **Version Compatibility**: Ensures CLI and plugin version compatibility
- **GitHub Downloads**: Install plugins directly from GitHub releases
- **gRPC Communication**: Uses gRPC for efficient plugin communication
- **Auto-Registration**: Plugins are automatically discovered and registered
- **Cobra CLI**: Nested command structure for extensibility

## Architecture

```
├── cmd/cli/           # Main CLI application
├── plugins/           # Plugin implementations
│   └── dummy/         # Example plugin
├── shared/            # Shared interfaces and protocols
├── pkg/              
│   ├── plugin/        # Plugin loading and management
│   ├── discovery/     # Plugin discovery logic
│   └── manager/       # Package manager for GitHub downloads
└── internal/
    ├── version/       # Version compatibility checking
    └── config/        # Configuration management
```

## Quick Start

### Prerequisites

- Go 1.22 or higher
- Protocol Buffers compiler (protoc)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/williamokano/hashicorp-plugin-example
cd hashicorp-plugin-example
```

2. Install dependencies:
```bash
make deps
```

3. Build everything:
```bash
make build
```

4. Install CLI and plugins:
```bash
make install
```

## Usage

### List Available Plugins
```bash
plugin-cli plugin list
```

### Get Plugin Information
```bash
plugin-cli plugin info dummy
```

### Run a Plugin
```bash
plugin-cli run -p dummy -m "Hello, World!"
```

### Install Plugin from GitHub
```bash
plugin-cli install owner/repo --version v1.0.0
```

### Remove a Plugin
```bash
plugin-cli plugin remove dummy
```

## Creating Your Own Plugin

1. Implement the `VersionedPlugin` interface:

```go
package main

import (
    "context"
    "github.com/williamokano/hashicorp-plugin-example/shared"
    "github.com/hashicorp/go-plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Process(ctx context.Context, input shared.DummyModel) (shared.DummyModel, error) {
    // Your plugin logic here
    return shared.DummyModel{
        Message: "Processed: " + input.Message,
    }, nil
}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) BuildTime() string { return "2024-01-01" }
func (p *MyPlugin) MinCLIVersion() string { return "1.0.0" }
func (p *MyPlugin) MaxCLIVersion() string { return "2.0.0" }
func (p *MyPlugin) Description() string { return "My custom plugin" }

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: shared.Handshake,
        Plugins: map[string]plugin.Plugin{
            "plugin": &shared.GRPCPlugin{Impl: &MyPlugin{}},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

2. Build your plugin:
```bash
go build -o plugin-myplugin myplugin.go
```

3. Place it in one of the plugin directories:
- `~/.local/share/plugins/`
- `./plugins/`
- Any path in `PLUGIN_PATH` environment variable

## Plugin Discovery Paths

The CLI searches for plugins in the following locations (in order):

1. `~/.local/share/plugins/`
2. `./plugins/` (current directory)
3. `./.plugins/` (hidden directory)
4. Paths specified in `PLUGIN_PATH` environment variable
5. `/usr/local/lib/plugins/`

Plugins must:
- Have executable permissions
- Start with `plugin-` prefix
- Implement the required interfaces

## Environment Variables

- `PLUGIN_PATH`: Colon-separated list of additional plugin directories
- `PLUGIN_LOG_LEVEL`: Control plugin system logging (default: `error`)
  - Options: `trace`, `debug`, `info`, `warn`, `error`, `off`
  - Example: `PLUGIN_LOG_LEVEL=debug plugin-cli plugin list`

## Configuration

Create a configuration file at `~/.config/plugin-cli/config.json`:

```json
{
  "plugins": [
    {
      "name": "dummy",
      "repository": "williamokano/plugin-dummy",
      "version": "v1.0.0",
      "enabled": true
    }
  ],
  "plugin_paths": [
    "~/.local/share/plugins",
    "./plugins"
  ],
  "auto_download": true
}
```

## Building from Source

### Generate Protocol Buffers
```bash
make proto
```

### Build CLI Only
```bash
make build-cli
```

### Build Plugins Only
```bash
make build-dummy-plugin
```

### Run Tests
```bash
make test
```

### Clean Build Artifacts
```bash
make clean
```

## Development

### Project Structure

- **shared/**: Contains the plugin interface and gRPC protocol definitions
- **pkg/plugin/**: Plugin manager for loading and executing plugins
- **pkg/discovery/**: Auto-discovery logic for finding plugins
- **pkg/manager/**: GitHub release downloader and installer
- **internal/version/**: Version compatibility checking
- **internal/config/**: Configuration file management

### Adding New Commands

Use Cobra to add new commands to the CLI:

```go
func newMyCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "mycommand",
        Short: "Description of my command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Command logic here
            return nil
        },
    }
}
```

## License

MIT

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.