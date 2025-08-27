# Project Structure

## Overview

The project follows a clean, modular architecture with clear separation of concerns. Each package has a specific responsibility, making the codebase maintainable and extensible.

```
hashicorp-plugin-example/
├── cmd/                      # Application entry points
│   └── cli/                  # CLI application
│       └── main.go          # Main entry point with cobra commands
│
├── plugins/                  # Plugin implementations
│   ├── dummy/               # Example dummy plugin
│   │   └── main.go
│   ├── filter/              # Message filtering plugin
│   │   └── main.go
│   ├── converter/           # Media conversion plugin
│   │   └── main.go
│   └── uploader/            # File upload plugin
│       └── main.go
│
├── pkg/                      # Core packages (public)
│   ├── types/               # Core type definitions
│   │   ├── event.go        # Event types and structures
│   │   ├── context.go      # Context and response types
│   │   └── plugin.go       # Plugin interfaces
│   │
│   ├── protocol/            # gRPC protocol implementation
│   │   ├── plugin.proto    # Protocol buffer definitions
│   │   ├── plugin.pb.go    # Generated protobuf code
│   │   ├── plugin_grpc.pb.go # Generated gRPC code
│   │   ├── handshake.go    # Plugin handshake config
│   │   ├── grpc_plugin.go  # gRPC plugin wrapper
│   │   ├── grpc_server.go  # gRPC server implementation
│   │   ├── grpc_client.go  # gRPC client implementation
│   │   └── converter.go    # Proto <-> Go type converters
│   │
│   ├── plugin/              # Plugin management
│   │   └── manager.go      # Plugin loading and lifecycle
│   │
│   ├── pipeline/            # Event processing pipeline
│   │   └── pipeline.go     # Pipeline orchestration
│   │
│   ├── discovery/           # Plugin discovery
│   │   └── discovery.go    # File system plugin discovery
│   │
│   └── manager/             # Package management
│       └── package.go      # GitHub plugin downloads
│
├── internal/                 # Internal packages (private)
│   ├── version/             # Version management
│   │   └── version.go      # Version compatibility checking
│   │
│   └── config/              # Configuration
│       └── config.go       # Config file handling
│
├── shared/                   # Backward compatibility layer
│   └── shared.go            # Re-exports for compatibility
│
├── docs/                     # Documentation
│   ├── architecture.md      # Architecture documentation
│   └── project-structure.md # This file
│
├── bin/                      # Build output (gitignored)
│   ├── plugin-cli           # CLI binary
│   └── plugin-*             # Plugin binaries
│
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── Makefile                  # Build automation
├── README.md                 # Project documentation
├── .gitignore               # Git ignore rules
└── plugin-config.example.json # Example configuration

```

## Package Descriptions

### `/cmd/cli`
**Purpose**: CLI application entry point  
**Responsibilities**:
- Command-line interface using Cobra
- User interaction and output formatting
- Command routing to appropriate handlers

### `/pkg/types`
**Purpose**: Core type definitions  
**Responsibilities**:
- Define Event, Context, Response structures
- Define Plugin interfaces
- Shared data structures

### `/pkg/protocol`
**Purpose**: gRPC communication protocol  
**Responsibilities**:
- Protocol buffer definitions
- gRPC server/client implementation
- Type conversion between Go and protobuf
- Plugin handshake configuration

### `/pkg/plugin`
**Purpose**: Plugin lifecycle management  
**Responsibilities**:
- Load plugins from filesystem
- Version compatibility checking
- Plugin client management
- Metadata retrieval

### `/pkg/pipeline`
**Purpose**: Event processing orchestration  
**Responsibilities**:
- Load and sort plugins by priority
- Execute plugins in sequence
- Pass context between plugins
- Handle plugin failures gracefully

### `/pkg/discovery`
**Purpose**: Plugin discovery  
**Responsibilities**:
- Scan filesystem for plugin binaries
- Search multiple configured paths
- Filter by naming convention (plugin-*)

### `/pkg/manager`
**Purpose**: Remote plugin management  
**Responsibilities**:
- Download plugins from GitHub
- Extract and install archives
- List and remove installed plugins

### `/internal/version`
**Purpose**: Version compatibility  
**Responsibilities**:
- Semantic version parsing
- Compatibility checking
- Version comparison

### `/internal/config`
**Purpose**: Configuration management  
**Responsibilities**:
- Load/save configuration files
- Default configuration values
- Plugin registry management

### `/shared`
**Purpose**: Backward compatibility  
**Responsibilities**:
- Re-export types from new locations
- Maintain API compatibility
- Migration path for existing code

## Design Principles

### 1. **Separation of Concerns**
Each package has a single, well-defined responsibility. Types are separated from protocol, which is separated from business logic.

### 2. **Clean Dependencies**
- `pkg/types` has no dependencies on other packages
- `pkg/protocol` only depends on `pkg/types`
- Higher-level packages depend on lower-level ones
- No circular dependencies

### 3. **Public vs Internal**
- `pkg/` contains packages that could be imported by plugins or external code
- `internal/` contains implementation details not meant for external use

### 4. **Interface-Driven Design**
- Core functionality defined through interfaces (`Plugin`, `VersionedPlugin`)
- Implementations can be swapped without changing consumers
- Testability through mock implementations

### 5. **Protocol Isolation**
- gRPC protocol details isolated in `pkg/protocol`
- Rest of the code works with Go types, not protobuf
- Protocol can be changed without affecting business logic

## Data Flow

### 1. **Event Reception**
```
User Input → CLI → Event Creation → Pipeline
```

### 2. **Plugin Processing**
```
Pipeline → Discovery → Load Plugins → Sort by Priority → Execute
```

### 3. **Context Enrichment**
```
Plugin 1 → Modify Context → Plugin 2 → Modify Context → Plugin 3
```

### 4. **Response Aggregation**
```
Each Plugin → Add Response → Context → Final Output → User
```

## Key Files

### Configuration Files
- `go.mod` - Go module dependencies
- `Makefile` - Build and development tasks
- `plugin-config.example.json` - Example plugin configuration

### Generated Files
- `pkg/protocol/plugin.pb.go` - Generated from proto
- `pkg/protocol/plugin_grpc.pb.go` - Generated gRPC code

### Entry Points
- `cmd/cli/main.go` - CLI application
- `plugins/*/main.go` - Individual plugin entry points

## Build System

### Makefile Targets
- `make proto` - Generate protobuf code
- `make build` - Build CLI and all plugins
- `make build-cli` - Build just the CLI
- `make build-plugins` - Build all plugins
- `make install` - Install CLI and plugins
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make deps` - Install dependencies

### Build Output
All binaries are placed in `bin/` directory:
- `bin/plugin-cli` - Main CLI executable
- `bin/plugin-*` - Plugin executables

## Development Workflow

1. **Adding a New Plugin**
   - Create new directory under `plugins/`
   - Implement `types.VersionedPlugin` interface
   - Add build target to Makefile
   - Test with CLI

2. **Modifying Core Types**
   - Edit types in `pkg/types/`
   - Update proto definitions if needed
   - Regenerate proto with `make proto`
   - Update implementations

3. **Adding CLI Commands**
   - Add new command function in `cmd/cli/main.go`
   - Implement command logic
   - Test with example data

## Testing Strategy

### Unit Tests
- Each package should have its own test file
- Mock interfaces for testing
- Focus on business logic

### Integration Tests
- Test plugin loading and execution
- Test pipeline processing
- Test CLI commands

### Plugin Tests
- Each plugin tested independently
- Mock context for testing
- Verify proper context modification

## Best Practices

1. **Always regenerate proto** after modifying `.proto` files
2. **Keep plugins simple** - one responsibility per plugin
3. **Document context properties** that plugins add/expect
4. **Version plugins properly** for compatibility
5. **Handle errors gracefully** - don't panic
6. **Log appropriately** - use structured logging
7. **Clean up resources** - close connections, kill processes

## Future Improvements

1. **Plugin Dependencies** - Allow plugins to declare dependencies
2. **Plugin Configuration** - Per-plugin configuration files
3. **Plugin Testing Framework** - Standardized plugin testing
4. **Plugin Marketplace** - Central registry for plugins
5. **Hot Reload** - Reload plugins without restart
6. **Metrics & Monitoring** - Track plugin performance
7. **Plugin Composition** - Combine plugins into workflows
8. **WebAssembly Support** - Run plugins in WASM sandbox

---

This structure provides a solid foundation for a plugin-based architecture that is:
- **Maintainable**: Clear separation and single responsibility
- **Extensible**: Easy to add new plugins and features
- **Testable**: Interface-driven design enables testing
- **Scalable**: Can handle many plugins efficiently
- **Professional**: Follows Go best practices and conventions