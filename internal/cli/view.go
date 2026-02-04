package cli

import "github.com/spf13/cobra"

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Browse and view XBE content",
	Long: `Browse and view XBE content.

The view command provides read-only access to XBE platform data. All view
commands support common flags documented in 'xbe --help'.

List/show commands also support:
  --fields     Sparse fieldset selection (list/show only)
  --omit-null  Omit null values in JSON output (list/show only)
  --version-changes  Include version history (versioned resources only)

Tip: Use 'xbe knowledge resources --version-changes' to list resources that support version history.
Tip: Optional feature gates for version history are auto-applied. See 'xbe knowledge resource <name>'.

Fields usage:
  --fields name,broker
  List default: label fields (or ID only). Show default: all fields.
  Relationships add <rel>-id automatically.`,
	Example: `  xbe view projects list                     # List all
  xbe view projects list --status active     # Filter
  xbe view projects show 123                 # Show one
  xbe view projects show 123 --version-changes
  xbe view projects list --json              # JSON output`,
	Annotations: map[string]string{"group": GroupCore},
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		return applyVersionChangesContext(cmd)
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.PersistentFlags().Bool("version-changes", false, "Include version changes in responses (supported resources only)")
	viewCmd.PersistentFlags().Bool("client-url", false, "Output client app URL(s) only")
}
