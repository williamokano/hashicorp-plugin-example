# Claude AI Assistant Instructions

## Project Context

This is a Go-based plugin architecture system inspired by HashiCorp's plugin system (used in Terraform, Vault, etc.). The project demonstrates event-driven plugin processing with auto-discovery, version compatibility, and GitHub-based distribution.

## Technical Stack

- **Language**: Go 1.22+
- **Protocol**: gRPC with Protocol Buffers
- **CLI Framework**: Cobra
- **Testing**: Testify + Counterfeiter
- **Plugin System**: HashiCorp go-plugin
- **Build Tool**: Make

## Project Structure

```
hashicorp-plugin-example/
├── cmd/cli/                 # CLI application
│   ├── main.go             # Entry point (minimal)
│   └── commands/           # One file per command
├── pkg/                     # Public packages
│   ├── types/              # Core type definitions
│   ├── protocol/           # gRPC protocol implementation
│   ├── plugin/             # Plugin management
│   ├── pipeline/           # Event processing pipeline
│   ├── discovery/          # Plugin discovery
│   ├── manager/            # Package management
│   └── interfaces/         # Interfaces for testing
├── internal/               # Private packages
│   ├── version/           # Version management
│   └── config/            # Configuration
├── plugins/               # Plugin implementations
│   ├── dummy/            # Example plugin
│   ├── filter/           # Message filter plugin
│   ├── converter/        # Media converter plugin
│   └── uploader/         # File uploader plugin
├── shared/               # Backward compatibility layer
└── docs/                 # Documentation
```

## Development Standards

### Code Organization

1. **Package Structure**:
   - Each package should have a single, clear responsibility
   - Interfaces go in `pkg/interfaces/` for mockability
   - Keep files under 300 lines when possible
   - One struct/interface per file for complex types

2. **File Naming**:
   - Use descriptive, lowercase names: `discovery.go`, `manager.go`
   - Test files: `<name>_test.go`
   - Mocks: `interfacesfakes/fake_<interface>.go` (generated)

3. **Command Structure**:
   - Each cobra command in its own file under `cmd/cli/commands/`
   - Subcommands should be functions within the parent command file
   - Command files should be 100-200 lines max

### Go Best Practices

1. **Error Handling**:
   ```go
   // Always wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to load plugin: %w", err)
   }
   ```

2. **Interface Design**:
   ```go
   // Keep interfaces small and focused
   type PluginLoader interface {
       LoadPlugin(name string) (*Plugin, error)
   }
   ```

3. **Naming Conventions**:
   - Interfaces: `<Name>er` or descriptive noun
   - Implementations: Concrete name
   - Test helpers: `Test<Name>` or `test<Name>`
   - Mock/Fake: `Fake<Interface>` (generated)

4. **Comments**:
   - Package comments: Start with `// Package <name>`
   - Exported types/functions: Start with `// <Name>`
   - Use complete sentences

### Testing Standards

1. **Test Structure**:
   ```go
   func TestFunctionName(t *testing.T) {
       // Arrange
       // Act  
       // Assert
   }
   ```

2. **Table-Driven Tests**:
   ```go
   tests := []struct {
       name    string
       input   string
       want    string
       wantErr bool
   }{
       // test cases
   }
   ```

3. **Using Testify**:
   ```go
   assert.Equal(t, expected, actual, "descriptive message")
   require.NoError(t, err) // Use require for critical assertions
   ```

4. **Using Counterfeiter**:
   ```go
   //go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
   //counterfeiter:generate . InterfaceName
   ```

5. **Test Coverage**:
   - Aim for 80%+ coverage on business logic
   - 100% coverage on critical paths
   - Skip coverage on generated code

### Plugin Development

1. **Plugin Interface Implementation**:
   - Implement `types.VersionedPlugin` interface
   - Priority: 0-100 (lower runs first)
   - Return proper `ExecutionDecision` in `ShouldExecute`

2. **Context Properties**:
   - Namespace your properties: `<plugin>_<property>`
   - Document expected inputs/outputs
   - Check types when reading: `val, ok := ctx.Properties["key"].(type)`

3. **Error Handling in Plugins**:
   - Never panic - return errors
   - Log errors appropriately
   - Continue processing on non-critical errors

### Protobuf/gRPC

1. **Updating Proto Files**:
   ```bash
   # After editing .proto files
   make proto
   ```

2. **Proto Location**: `pkg/protocol/plugin.proto`

3. **Generated Files**: Don't edit `*.pb.go` files

### Build & Release

1. **Building**:
   ```bash
   make build         # Build everything
   make build-cli     # Build CLI only
   make build-plugins # Build all plugins
   ```

2. **Testing**:
   ```bash
   make test          # Run all tests
   make test-coverage # Run with coverage
   ```

3. **Installation**:
   ```bash
   make install       # Install CLI and plugins
   ```

## Command Reference

### Main Commands
- `plugin-cli version` - Version information
- `plugin-cli plugin [subcommand]` - Plugin management
- `plugin-cli process [message]` - Process events
- `plugin-cli install [repo]` - Install from GitHub
- `plugin-cli simulate [scenario]` - Run simulations

### Common Tasks

1. **Adding a New Command**:
   - Create file in `cmd/cli/commands/`
   - Add to `root.go` command list
   - Follow existing command patterns

2. **Adding a New Plugin**:
   - Create directory under `plugins/`
   - Implement `types.VersionedPlugin`
   - Add to Makefile build targets

3. **Adding a New Package**:
   - Create under `pkg/` (public) or `internal/` (private)
   - Add interface to `pkg/interfaces/` if mockable
   - Write tests alongside implementation

## Testing Instructions

### Running Tests
```bash
# All tests
make test

# Specific package
go test ./pkg/discovery/...

# With coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Generating Mocks
```bash
# Install counterfeiter
go install github.com/maxbrunsfeld/counterfeiter/v6@latest

# Generate all mocks
go generate ./...

# Generate specific interface
counterfeiter -o pkg/interfaces/fakes/fake_plugin_manager.go pkg/interfaces PluginManager
```

### Writing Tests

1. **Unit Test Example**:
```go
func TestDiscoverPlugins(t *testing.T) {
    // Setup
    tempDir := t.TempDir()
    createTestPlugin(t, tempDir, "plugin-test")
    
    // Execute
    plugins, err := discovery.DiscoverPlugins([]string{tempDir})
    
    // Assert
    require.NoError(t, err)
    assert.Len(t, plugins, 1)
    assert.Equal(t, "test", plugins[0].Name)
}
```

2. **Integration Test Example**:
```go
func TestPipelineExecution(t *testing.T) {
    // Use real components, not mocks
    pipeline := pipeline.NewPipeline()
    ctx, err := pipeline.ProcessMessage(context.Background(), 
        "test", "message", "user", "channel")
    
    require.NoError(t, err)
    assert.NotNil(t, ctx)
}
```

## Important Implementation Details

1. **Plugin Communication**: Uses gRPC over local Unix sockets
2. **Plugin Discovery**: Searches for `plugin-*` binaries in configured paths
3. **Version Checking**: Semantic versioning with min/max CLI version
4. **Context Passing**: Plugins share data via `Context.Properties` map
5. **Priority Execution**: Plugins run in priority order (0-100)

## Common Issues & Solutions

1. **Plugin Not Found**:
   - Check binary name starts with `plugin-`
   - Verify executable permissions
   - Check discovery paths

2. **Version Incompatibility**:
   - Check plugin's min/max version requirements
   - Update CLI or plugin as needed

3. **gRPC Errors**:
   - Regenerate proto files: `make proto`
   - Check handshake configuration matches

## Environment Variables

- `PLUGIN_PATH`: Additional plugin discovery paths (colon-separated)
- `PLUGIN_LOG_LEVEL`: Log verbosity (debug, info, warn, error)

## Dependencies

Key dependencies (see go.mod for versions):
- github.com/hashicorp/go-plugin
- github.com/spf13/cobra
- google.golang.org/grpc
- github.com/stretchr/testify
- github.com/maxbrunsfeld/counterfeiter/v6

## Contribution Workflow

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Write tests first (TDD preferred)
4. Implement feature
5. Ensure tests pass: `make test`
6. Update documentation if needed
7. Submit pull request

## Code Review Checklist

- [ ] Tests written and passing
- [ ] Interfaces created for mockable components
- [ ] Error handling with proper context
- [ ] Documentation updated
- [ ] No lint errors
- [ ] Coverage maintained or improved

## Notes for Claude

When working on this project:
1. Always create interfaces for testable components
2. Write tests alongside implementation
3. Use table-driven tests for multiple scenarios
4. Keep commands in separate files
5. Follow the established package structure
6. Use proper error wrapping with context
7. Generate mocks with counterfeiter
8. Maintain backward compatibility in `shared/` package
9. Update this file when adding major features
10. Prefer composition over inheritance

## Release Checklist

- [ ] All tests passing
- [ ] Version bumped in `internal/version/version.go`
- [ ] CHANGELOG.md updated
- [ ] Documentation updated
- [ ] Proto files regenerated if changed
- [ ] Binaries built for all platforms
- [ ] GitHub release created with binaries