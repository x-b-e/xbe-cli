package cli

import "github.com/spf13/cobra"

var doOpenAiRealtimeSessionsCmd = &cobra.Command{
	Use:     "open-ai-realtime-sessions",
	Aliases: []string{"open-ai-realtime-session"},
	Short:   "Manage OpenAI realtime sessions",
	Long:    "Create OpenAI realtime sessions.",
}

func init() {
	doCmd.AddCommand(doOpenAiRealtimeSessionsCmd)
}
