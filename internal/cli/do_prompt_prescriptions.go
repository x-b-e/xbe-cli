package cli

import "github.com/spf13/cobra"

var doPromptPrescriptionsCmd = &cobra.Command{
	Use:     "prompt-prescriptions",
	Aliases: []string{"prompt-prescription"},
	Short:   "Manage prompt prescriptions",
	Long: `Submit prompt prescription requests and retrieve generated prompts.

Commands:
  create    Create a prompt prescription request`,
	Example: `  # Create a prompt prescription
  xbe do prompt-prescriptions create \\
    --email-address "name@example.com" \\
    --name "Alex Builder" \\
    --organization-name "Concrete Co" \\
    --location-name "Austin, TX" \\
    --role "Operations Manager" \\
    --symptoms "Rising costs, scheduling delays"`,
}

func init() {
	doCmd.AddCommand(doPromptPrescriptionsCmd)
}
