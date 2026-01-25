package cli

import "github.com/spf13/cobra"

var doMaterialSiteMergersCmd = &cobra.Command{
	Use:     "material-site-mergers",
	Aliases: []string{"material-site-merger"},
	Short:   "Merge material sites",
	Long: `Merge one material site into another.

Merging moves references from the orphan material site to the survivor and
removes the orphan. This action is destructive and requires admin access.

Commands:
  create    Merge a material site`,
	Example: `  # Merge an orphan material site into a survivor
  xbe do material-site-mergers create --orphan 123 --survivor 456`,
}

func init() {
	doCmd.AddCommand(doMaterialSiteMergersCmd)
}
