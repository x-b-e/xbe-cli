package cli

import "github.com/spf13/cobra"

var siteWaitTimeNotificationTriggersCmd = &cobra.Command{
	Use:     "site-wait-time-notification-triggers",
	Aliases: []string{"site-wait-time-notification-trigger"},
	Short:   "Browse site wait time notification triggers",
	Long: `Browse site wait time notification triggers.

Site wait time notification triggers record excessive wait time events for job sites
and material sites that generate notifications.

Commands:
  list  List site wait time notification triggers with filtering and pagination
  show  Show site wait time notification trigger details`,
	Example: `  # List site wait time notification triggers
  xbe view site-wait-time-notification-triggers list

  # Filter by job production plan
  xbe view site-wait-time-notification-triggers list --job-production-plan 123

  # Filter by site type
  xbe view site-wait-time-notification-triggers list --site-type job_site

  # Show trigger details
  xbe view site-wait-time-notification-triggers show 456`,
}

func init() {
	viewCmd.AddCommand(siteWaitTimeNotificationTriggersCmd)
}
