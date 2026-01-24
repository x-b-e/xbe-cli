package cli

import "github.com/spf13/cobra"

var doSourcingSearchesCmd = &cobra.Command{
	Use:     "sourcing-searches",
	Aliases: []string{"sourcing-search"},
	Short:   "Find matching truckers and trailers for a customer tender",
	Long: `Run a sourcing search to find truckers, trailers, and broker tenders that match
a customer tender.

Commands:
  create    Run a sourcing search`,
	Example: `  # Run a sourcing search with defaults
  xbe do sourcing-searches create --customer-tender 123

  # Constrain the search
  xbe do sourcing-searches create --customer-tender 123 \
    --maximum-distance-miles 75 --maximum-result-count 25

  # Require additional certification types
  xbe do sourcing-searches create --customer-tender 123 \
    --additional-certification-requirement-types 45,67

  # JSON output
  xbe do sourcing-searches create --customer-tender 123 --json`,
}

func init() {
	doCmd.AddCommand(doSourcingSearchesCmd)
}
