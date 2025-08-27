package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/pipeline"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// NewSimulateCommand creates the simulate command group
func NewSimulateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Run example simulations",
		Long: `Run predefined simulation scenarios to test the plugin pipeline.

Available simulations demonstrate different use cases and plugin interactions.`,
	}

	cmd.AddCommand(
		newSimulateDiscordCommand(),
		newSimulateVideoCommand(),
		newSimulateTelegramCommand(),
		newSimulateWorkflowCommand(),
	)

	return cmd
}

func newSimulateDiscordCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "discord",
		Short: "Simulate a Discord message",
		Long:  `Simulate a Discord user sending a message requesting video conversion and upload.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🎮 Simulating Discord message...")
			fmt.Println("User: discord_user_123")
			fmt.Println("Channel: #general")
			fmt.Println("Message: 'Please convert this video and upload it'")
			fmt.Println()

			p := pipeline.NewPipeline()
			ctx, err := p.ProcessMessage(
				context.Background(),
				"discord",
				"Please convert this video and upload it",
				"discord_user_123",
				"general",
			)

			if err != nil {
				return fmt.Errorf("simulation failed: %w", err)
			}

			printSimulationResult(ctx)
			return nil
		},
	}
}

func newSimulateVideoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "video",
		Short: "Simulate a video processing request",
		Long:  `Simulate a complete video processing workflow with file attachment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🎥 Simulating video processing request...")
			fmt.Println("Source: Telegram DM")
			fmt.Println("Command: 'convert video to mp4 and upload to cloud'")
			fmt.Println("Attachment: video.mov")
			fmt.Println()

			p := pipeline.NewPipeline()

			event := types.Event{
				Type:      types.EventCommand,
				Source:    "telegram",
				Content:   "convert video to mp4 and upload to cloud",
				UserID:    "telegram_user_456",
				ChannelID: "dm_789",
				Metadata: map[string]interface{}{
					"has_attachment": true,
					"file_url":       "https://example.com/video.mov",
					"file_size":      1024 * 1024 * 50, // 50MB
					"file_type":      "video/quicktime",
				},
			}

			ctx, err := p.ProcessEvent(context.Background(), event)
			if err != nil {
				return fmt.Errorf("simulation failed: %w", err)
			}

			printSimulationResult(ctx)
			return nil
		},
	}
}

func newSimulateTelegramCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "telegram",
		Short: "Simulate a Telegram bot command",
		Long:  `Simulate a Telegram bot receiving a command with multiple processing steps.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("💬 Simulating Telegram bot command...")
			fmt.Println("User: @john_doe")
			fmt.Println("Command: /process")
			fmt.Println("Arguments: image enhance, resize, upload")
			fmt.Println()

			p := pipeline.NewPipeline()

			event := types.Event{
				Type:      types.EventCommand,
				Source:    "telegram",
				Content:   "/process image enhance resize upload",
				UserID:    "@john_doe",
				ChannelID: "private_chat",
				Metadata: map[string]interface{}{
					"command":       "/process",
					"args":          []string{"image", "enhance", "resize", "upload"},
					"message_id":    123456,
					"reply_to":      123455,
					"has_media":     true,
					"media_type":    "photo",
					"media_file_id": "AgACAgIAAxkBAAI...",
				},
			}

			ctx, err := p.ProcessEvent(context.Background(), event)
			if err != nil {
				return fmt.Errorf("simulation failed: %w", err)
			}

			printSimulationResult(ctx)
			return nil
		},
	}
}

func newSimulateWorkflowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "workflow",
		Short: "Simulate a complex multi-step workflow",
		Long: `Simulate a complex workflow that triggers multiple plugins in sequence,
demonstrating context enrichment and plugin cooperation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⚙️ Simulating complex workflow...")
			fmt.Println("Workflow: Media Processing Pipeline")
			fmt.Println("Steps: Download → Validate → Convert → Optimize → Upload → Notify")
			fmt.Println()

			p := pipeline.NewPipeline()

			event := types.Event{
				Type:      types.EventWebhook,
				Source:    "api",
				Content:   "process media workflow",
				UserID:    "system",
				ChannelID: "webhook",
				Metadata: map[string]interface{}{
					"workflow_id": "wf_123",
					"priority":    "high",
					"steps": []string{
						"download",
						"validate",
						"convert",
						"optimize",
						"upload",
						"notify",
					},
					"source_url":     "https://source.example.com/media.raw",
					"target_format":  "mp4",
					"target_quality": "1080p",
					"notify_url":     "https://webhook.example.com/complete",
				},
			}

			ctx, err := p.ProcessEvent(context.Background(), event)
			if err != nil {
				return fmt.Errorf("simulation failed: %w", err)
			}

			printSimulationResult(ctx)

			// Show workflow summary
			fmt.Println("\n📊 Workflow Summary:")
			fmt.Printf("  Total Plugins Executed: %d\n", len(ctx.Responses))
			fmt.Printf("  Properties Added: %d\n", len(ctx.Properties))

			if uploadURL, ok := ctx.Properties["uploaded_url"]; ok {
				fmt.Printf("  Final Output: %v\n", uploadURL)
			}

			return nil
		},
	}
}

func printSimulationResult(ctx *types.Context) {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║      SIMULATION RESULT               ║")
	fmt.Println("╚══════════════════════════════════════╝")

	fmt.Printf("\n📨 Event Details:\n")
	fmt.Printf("   Type: %s\n", ctx.Event.Type)
	fmt.Printf("   Source: %s\n", ctx.Event.Source)
	fmt.Printf("   Content: %s\n", ctx.Event.Content)

	if len(ctx.Event.Metadata) > 0 {
		fmt.Printf("\n📎 Metadata:\n")
		for k, v := range ctx.Event.Metadata {
			fmt.Printf("   • %s: %v\n", k, v)
		}
	}

	if len(ctx.Properties) > 0 {
		fmt.Printf("\n🔧 Context Properties:\n")
		for k, v := range ctx.Properties {
			// Skip complex objects for cleaner output
			if _, ok := v.(map[string]interface{}); !ok {
				fmt.Printf("   • %s: %v\n", k, v)
			}
		}
	}

	if len(ctx.Responses) > 0 {
		fmt.Printf("\n📤 Plugin Responses:\n")
		for i, resp := range ctx.Responses {
			fmt.Printf("   %d. [%s] %s\n", i+1, resp.PluginName, resp.Content)
		}
	}

	fmt.Println()
}
