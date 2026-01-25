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

type doJobProductionPlanMaterialSiteChangesCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	JobProductionPlan    string
	OldMaterialSite      string
	NewMaterialSite      string
	OldMaterialType      string
	NewMaterialType      string
	NewMaterialMixDesign string
}

func newDoJobProductionPlanMaterialSiteChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan material site change",
		Long: `Create a job production plan material site change.

Required flags:
  --job-production-plan  Job production plan ID (required)
  --old-material-site    Old material site ID (required)
  --new-material-site    New material site ID (required)

Optional flags:
  --old-material-type      Old material type ID to swap (optional)
  --new-material-type      New material type ID (optional)
  --new-material-mix-design New material mix design ID (optional; requires --new-material-type)

Notes:
  The old material site must already be part of the job production plan.
  The old and new material sites must be different and belong to the plan's broker.
  Material type swaps are optional but must be different when both are set.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Swap a material site on a job production plan
  xbe do job-production-plan-material-site-changes create \
    --job-production-plan 123 \
    --old-material-site 456 \
    --new-material-site 789

  # Swap material site and material type
  xbe do job-production-plan-material-site-changes create \
    --job-production-plan 123 \
    --old-material-site 456 \
    --new-material-site 789 \
    --old-material-type 111 \
    --new-material-type 222`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanMaterialSiteChangesCreate,
	}
	initDoJobProductionPlanMaterialSiteChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialSiteChangesCmd.AddCommand(newDoJobProductionPlanMaterialSiteChangesCreateCmd())
}

func initDoJobProductionPlanMaterialSiteChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("old-material-site", "", "Old material site ID (required)")
	cmd.Flags().String("new-material-site", "", "New material site ID (required)")
	cmd.Flags().String("old-material-type", "", "Old material type ID (optional)")
	cmd.Flags().String("new-material-type", "", "New material type ID (optional)")
	cmd.Flags().String("new-material-mix-design", "", "New material mix design ID (optional)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanMaterialSiteChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanMaterialSiteChangesCreateOptions(cmd)
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

	if opts.JobProductionPlan == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.OldMaterialSite == "" {
		err := fmt.Errorf("--old-material-site is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NewMaterialSite == "" {
		err := fmt.Errorf("--new-material-site is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NewMaterialMixDesign != "" && opts.NewMaterialType == "" {
		err := fmt.Errorf("--new-material-mix-design requires --new-material-type")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"old-material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.OldMaterialSite,
			},
		},
		"new-material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.NewMaterialSite,
			},
		},
	}

	if opts.OldMaterialType != "" {
		relationships["old-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.OldMaterialType,
			},
		}
	}
	if opts.NewMaterialType != "" {
		relationships["new-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.NewMaterialType,
			},
		}
	}
	if opts.NewMaterialMixDesign != "" {
		relationships["new-material-mix-design"] = map[string]any{
			"data": map[string]any{
				"type": "material-mix-designs",
				"id":   opts.NewMaterialMixDesign,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-material-site-changes",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-material-site-changes", jsonBody)
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

	row := buildJobProductionPlanMaterialSiteChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan material site change %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanMaterialSiteChangesCreateOptions(cmd *cobra.Command) (doJobProductionPlanMaterialSiteChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	oldMaterialSite, _ := cmd.Flags().GetString("old-material-site")
	newMaterialSite, _ := cmd.Flags().GetString("new-material-site")
	oldMaterialType, _ := cmd.Flags().GetString("old-material-type")
	newMaterialType, _ := cmd.Flags().GetString("new-material-type")
	newMaterialMixDesign, _ := cmd.Flags().GetString("new-material-mix-design")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialSiteChangesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		JobProductionPlan:    jobProductionPlan,
		OldMaterialSite:      oldMaterialSite,
		NewMaterialSite:      newMaterialSite,
		OldMaterialType:      oldMaterialType,
		NewMaterialType:      newMaterialType,
		NewMaterialMixDesign: newMaterialMixDesign,
	}, nil
}
