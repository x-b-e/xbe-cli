package cli

import "github.com/spf13/cobra"

var openAiRealtimeSessionsCmd = &cobra.Command{
	Use:     "open-ai-realtime-sessions",
	Aliases: []string{"open-ai-realtime-session"},
	Short:   "View OpenAI realtime sessions",
	Long: `Browse and inspect OpenAI realtime sessions.

OpenAI realtime sessions store short-lived client secrets for realtime
streaming connections.

Commands:
  list    List sessions with filtering and pagination
  show    View full session details`,
	Example: `  # List sessions
  xbe view open-ai-realtime-sessions list

  # View a session
  xbe view open-ai-realtime-sessions show 123`,
}

func init() {
	viewCmd.AddCommand(openAiRealtimeSessionsCmd)
}
