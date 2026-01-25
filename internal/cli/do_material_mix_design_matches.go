package cli

import "github.com/spf13/cobra"

var doMaterialMixDesignMatchesCmd = &cobra.Command{
	Use:     "material-mix-design-matches",
	Aliases: []string{"material-mix-design-match"},
	Short:   "Match material mix designs",
	Long: `Find material mix designs that match a material type and optional material sites.

Commands:
  create    Match material mix designs`,
	Example: `  # Match material mix designs for a material type
  xbe do material-mix-design-matches create --material-type 123 --as-of "2026-01-23T00:00:00Z"

  # Match with material sites
  xbe do material-mix-design-matches create \
    --material-type 123 \
    --as-of "2026-01-23T00:00:00Z" \
    --material-sites 456,789`,
}

func init() {
	doCmd.AddCommand(doMaterialMixDesignMatchesCmd)
}
