package cli

import "github.com/spf13/cobra"

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Create, update, and delete XBE resources",
	Long: `Create, update, and delete XBE resources.

The do command provides write access to XBE platform data. Unlike view commands,
these operations modify data and require authentication.`,
	Example: `  xbe do customers create --name "Acme Corp"      # Create
  xbe do projects update 123 --status complete    # Update
  xbe do posts delete 456 --confirm               # Delete`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(doCmd)
}
