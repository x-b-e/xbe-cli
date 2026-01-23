package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanUnscrappagesCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	JobProductionPlanID               string
	Comment                           string
	SuppressStatusChangeNotifications bool
	SkipValidateRequiredMixDesigns    bool
}

type jobProductionPlanUnscrappageRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id,omitempty"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
	SkipValidateRequiredMixDesigns    bool   `json:"skip_validate_required_mix_designs"`
}

func newDoJobProductionPlanUnscrappagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unscrap a job production plan",
		Long: `Unscrap a job production plan.

This action transitions the plan status to approved. Only plans in scrapped
status can be unscrapped.

Required flags:
  --job-production-plan   Job production plan ID

Optional flags:
  --comment                                 Comment for the unscrappage
  --suppress-status-change-notifications    Suppress status change notifications
  --skip-validate-required-mix-designs      Skip required material mix design validation`,
		Example: `  # Unscrap a job production plan
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --comment "Restoring plan"

  # Unscrap and suppress notifications
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --suppress-status-change-notifications

  # Skip required mix design validation
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --skip-validate-required-mix-designs

  # JSON output
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --json`,
		RunE: runDoJobProductionPlanUnscrappagesCreate,
	}
	initDoJobProductionPlanUnscrappagesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanUnscrappagesCmd.AddCommand(newDoJobProductionPlanUnscrappagesCreateCmd())
}

func initDoJobProductionPlanUnscrappagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("comment", "", "Comment for the unscrappage")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().Bool("skip-validate-required-mix-designs", false, "Skip required material mix design validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job-production-plan")
}

func runDoJobProductionPlanUnscrappagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanUnscrappagesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}
	if cmd.Flags().Changed("suppress-status-change-notifications") {
		attributes["suppress-status-change-notifications"] = opts.SuppressStatusChangeNotifications
	}
	if cmd.Flags().Changed("skip-validate-required-mix-designs") {
		attributes["skip-validate-required-mix-designs"] = opts.SkipValidateRequiredMixDesigns
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-unscrappages",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-unscrappages", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := buildJobProductionPlanUnscrappageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan unscrappage %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanUnscrappagesCreateOptions(cmd *cobra.Command) (doJobProductionPlanUnscrappagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	comment, _ := cmd.Flags().GetString("comment")
	suppressNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	skipValidateRequiredMixDesigns, _ := cmd.Flags().GetBool("skip-validate-required-mix-designs")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanUnscrappagesCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		JobProductionPlanID:               jobProductionPlanID,
		Comment:                           comment,
		SuppressStatusChangeNotifications: suppressNotifications,
		SkipValidateRequiredMixDesigns:    skipValidateRequiredMixDesigns,
	}, nil
}

func buildJobProductionPlanUnscrappageRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanUnscrappageRow {
	resource := resp.Data
	row := jobProductionPlanUnscrappageRow{
		ID:                                resource.ID,
		Comment:                           stringAttr(resource.Attributes, "comment"),
		SuppressStatusChangeNotifications: boolAttr(resource.Attributes, "suppress-status-change-notifications"),
		SkipValidateRequiredMixDesigns:    boolAttr(resource.Attributes, "skip-validate-required-mix-designs"),
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	return row
}
