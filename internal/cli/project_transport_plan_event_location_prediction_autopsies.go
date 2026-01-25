package cli

import "github.com/spf13/cobra"

var projectTransportPlanEventLocationPredictionAutopsiesCmd = &cobra.Command{
	Use:     "project-transport-plan-event-location-prediction-autopsies",
	Aliases: []string{"project-transport-plan-event-location-prediction-autopsy"},
	Short:   "View project transport plan event location prediction autopsies",
	Long: `View project transport plan event location prediction autopsies.

Location prediction autopsies capture diagnostics for project transport plan
event location predictions, including status, errors, context, and LLM output.

Commands:
  list  List project transport plan event location prediction autopsies
  show  Show project transport plan event location prediction autopsy details`,
	Example: `  # List recent autopsies
  xbe view project-transport-plan-event-location-prediction-autopsies list --limit 10

  # Show an autopsy
  xbe view project-transport-plan-event-location-prediction-autopsies show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanEventLocationPredictionAutopsiesCmd)
}
