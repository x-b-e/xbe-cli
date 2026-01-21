package cli

import "github.com/spf13/cobra"

var certificationsCmd = &cobra.Command{
	Use:     "certifications",
	Aliases: []string{"certification"},
	Short:   "View certifications",
	Long:    "Commands for viewing certifications.",
}

func init() {
	viewCmd.AddCommand(certificationsCmd)
}
