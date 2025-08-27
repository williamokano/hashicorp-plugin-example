package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/pipeline"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// ProcessFlags holds flags for the process command
type ProcessFlags struct {
	Source     string
	UserID     string
	ChannelID  string
	EventType  string
	Metadata   string
	OutputJSON bool
	Quiet      bool
}

// NewProcessCommand creates the process command
func NewProcessCommand() *cobra.Command {
	flags := &ProcessFlags{}

	cmd := &cobra.Command{
		Use:   "process [message]",
		Short: "Process a message or event through all plugins",
		Long: `Process an event through the plugin pipeline. Each plugin will:
1. Decide if it should execute based on the event
2. Process the event and potentially modify the context
3. Pass the enriched context to the next plugin

The final context with all modifications and responses is returned.`,
		Example: `  # Simple message
  plugin-cli process "Convert this video"

  # Discord message with metadata
  plugin-cli process "Upload the file" --source discord --user john123

  # Command with JSON metadata
  plugin-cli process "Process data" --type command --metadata '{"priority": "high"}'

  # Quiet mode (only show results)
  plugin-cli process "Test message" --quiet --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProcess(args[0], flags)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&flags.Source, "source", "s", "cli", "Event source (discord, telegram, slack, cli)")
	cmd.Flags().StringVarP(&flags.UserID, "user", "u", "user123", "User ID")
	cmd.Flags().StringVarP(&flags.ChannelID, "channel", "c", "channel456", "Channel ID")
	cmd.Flags().StringVarP(&flags.EventType, "type", "t", string(types.EventMessage), "Event type (message, command, webhook, scheduled)")
	cmd.Flags().StringVarP(&flags.Metadata, "metadata", "m", "", "Additional metadata as JSON")
	cmd.Flags().BoolVar(&flags.OutputJSON, "json", false, "Output result as JSON")
	cmd.Flags().BoolVarP(&flags.Quiet, "quiet", "q", false, "Suppress processing logs")

	return cmd
}

func runProcess(content string, flags *ProcessFlags) error {
	// Parse metadata if provided
	var metadataMap map[string]interface{}
	if flags.Metadata != "" {
		if err := json.Unmarshal([]byte(flags.Metadata), &metadataMap); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
	} else {
		metadataMap = make(map[string]interface{})
	}

	// Create event
	event := types.Event{
		Type:      types.EventType(flags.EventType),
		Source:    flags.Source,
		Content:   content,
		UserID:    flags.UserID,
		ChannelID: flags.ChannelID,
		Metadata:  metadataMap,
	}

	// Process through pipeline
	p := pipeline.NewPipeline()
	ctx, err := p.ProcessEvent(context.Background(), event)
	if err != nil {
		return fmt.Errorf("pipeline processing failed: %w", err)
	}

	// Output results
	if flags.OutputJSON {
		outputJSON(ctx)
	} else if !flags.Quiet {
		outputFormatted(ctx)
	} else {
		outputMinimal(ctx)
	}

	return nil
}

func outputJSON(ctx *types.Context) {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	fmt.Println(string(data))
}

func outputFormatted(ctx *types.Context) {
	fmt.Println("\n=== Pipeline Execution Complete ===")
	fmt.Printf("Event: %s from %s\n", ctx.Event.Type, ctx.Event.Source)
	fmt.Printf("Content: %s\n\n", ctx.Event.Content)

	if len(ctx.Properties) > 0 {
		fmt.Println("Context Properties:")
		for k, v := range ctx.Properties {
			fmt.Printf("  %s: %v\n", k, v)
		}
		fmt.Println()
	}

	if len(ctx.Responses) > 0 {
		fmt.Println("Plugin Responses:")
		for _, resp := range ctx.Responses {
			fmt.Printf("  [%s] %s: %s\n", resp.PluginName, resp.Type, resp.Content)
			if len(resp.Data) > 0 && isVerbose() {
				data, _ := json.MarshalIndent(resp.Data, "    ", "  ")
				fmt.Printf("    Data: %s\n", data)
			}
		}
	}
}

func outputMinimal(ctx *types.Context) {
	for _, resp := range ctx.Responses {
		fmt.Printf("[%s] %s\n", resp.PluginName, resp.Content)
	}
}

func isVerbose() bool {
	// Check if verbose flag is set (would need to be passed from root)
	return false
}
