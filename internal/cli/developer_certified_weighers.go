package cli

import "github.com/spf13/cobra"

var developerCertifiedWeighersCmd = &cobra.Command{
	Use:   "developer-certified-weighers",
	Short: "View developer certified weighers",
	Long: `View developer certified weighers.

Developer certified weighers link developers to users who are certified
for weighing materials.

Commands:
  list  List developer certified weighers
  show  Show developer certified weigher details`,
}

func init() {
	viewCmd.AddCommand(developerCertifiedWeighersCmd)
}
