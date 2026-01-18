package cli

import "github.com/spf13/cobra"

var postsCmd = &cobra.Command{
	Use:   "posts",
	Short: "Browse and view posts",
	Long: `Browse and view posts on the XBE platform.

Posts include various content types such as status updates, key result completions,
production recaps, and general announcements. You can list posts with various filters
or view the full content of a specific post.

Commands:
  list    List posts with filtering and pagination
  show    View the full content of a specific post

Post Types:
  basic, notification, action, new_membership, post_summary, post_activity,
  objective_status, objective_completion, objective_status_scoreboard,
  key_result_completion, key_result_status_scoreboard,
  job_production_plan_recap, customer_daily_job_production_plan_recap,
  customer_job_production_plan_schedule, customer_lineup_schedule,
  material_supplier_production_daily_recap, material_supplier_production_monthly_recap,
  trucker_shift_summary, trucking_time_card_administration_report_card,
  trucking_tender_acceptance_report_card, driver_day_recap, release_note_summary

Filtering:
  The list command supports filtering by:
  - Status (draft/published)
  - Post type
  - Publication date ranges
  - Creator`,
	Example: `  # List recent published posts
  xbe view posts list

  # Filter by status
  xbe view posts list --status published

  # Filter by post type
  xbe view posts list --post-type basic

  # Filter by date range
  xbe view posts list --published-at-min 2024-01-01 --published-at-max 2024-12-31

  # Get results as JSON for scripting
  xbe view posts list --json --limit 10

  # View a specific post
  xbe view posts show 456`,
}

func init() {
	viewCmd.AddCommand(postsCmd)
}
