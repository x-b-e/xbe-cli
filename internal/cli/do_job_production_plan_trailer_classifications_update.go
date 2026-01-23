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

type doJobProductionPlanTrailerClassificationsUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	JobProductionPlan                  string
	TrailerClassification              string
	TrailerClassificationEquivalentIDs []string
	GrossWeightLegalLimitLbsExplicit   string
	ExplicitMaterialTransactionTonsMax string
}

func newDoJobProductionPlanTrailerClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan trailer classification",
		Long: `Update a job production plan trailer classification.

Optional flags:
  --job-production-plan                   Job production plan ID
  --trailer-classification                Trailer classification ID
  --trailer-classification-equivalent-ids Equivalent trailer classification IDs (comma-separated or repeated)
  --gross-weight-legal-limit-lbs-explicit Explicit gross weight legal limit (lbs)
  --explicit-material-transaction-tons-max Explicit material transaction tons max

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update explicit limits
  xbe do job-production-plan-trailer-classifications update 123 \
    --gross-weight-legal-limit-lbs-explicit 90000 \
    --explicit-material-transaction-tons-max 22

  # Update equivalent trailer classifications
  xbe do job-production-plan-trailer-classifications update 123 \
    --trailer-classification-equivalent-ids 111,222

  # Update relationships
  xbe do job-production-plan-trailer-classifications update 123 \
    --job-production-plan 456 \
    --trailer-classification 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanTrailerClassificationsUpdate,
	}
	initDoJobProductionPlanTrailerClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanTrailerClassificationsCmd.AddCommand(newDoJobProductionPlanTrailerClassificationsUpdateCmd())
}

func initDoJobProductionPlanTrailerClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().StringSlice("trailer-classification-equivalent-ids", nil, "Equivalent trailer classification IDs (comma-separated or repeated)")
	cmd.Flags().String("gross-weight-legal-limit-lbs-explicit", "", "Explicit gross weight legal limit (lbs)")
	cmd.Flags().String("explicit-material-transaction-tons-max", "", "Explicit material transaction tons max")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanTrailerClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanTrailerClassificationsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("trailer-classification-equivalent-ids") {
		attributes["trailer-classification-equivalent-ids"] = opts.TrailerClassificationEquivalentIDs
	}
	if cmd.Flags().Changed("gross-weight-legal-limit-lbs-explicit") {
		attributes["gross-weight-legal-limit-lbs-explicit"] = opts.GrossWeightLegalLimitLbsExplicit
	}
	if cmd.Flags().Changed("explicit-material-transaction-tons-max") {
		attributes["explicit-material-transaction-tons-max"] = opts.ExplicitMaterialTransactionTonsMax
	}
	if cmd.Flags().Changed("job-production-plan") {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if cmd.Flags().Changed("trailer-classification") {
		relationships["trailer-classification"] = map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "job-production-plan-trailer-classifications",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-trailer-classifications/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanTrailerClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan trailer classification %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanTrailerClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanTrailerClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	trailerClassificationEquivalentIDs, _ := cmd.Flags().GetStringSlice("trailer-classification-equivalent-ids")
	grossWeightLegalLimitLbsExplicit, _ := cmd.Flags().GetString("gross-weight-legal-limit-lbs-explicit")
	explicitMaterialTransactionTonsMax, _ := cmd.Flags().GetString("explicit-material-transaction-tons-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanTrailerClassificationsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		JobProductionPlan:                  jobProductionPlan,
		TrailerClassification:              trailerClassification,
		TrailerClassificationEquivalentIDs: trailerClassificationEquivalentIDs,
		GrossWeightLegalLimitLbsExplicit:   grossWeightLegalLimitLbsExplicit,
		ExplicitMaterialTransactionTonsMax: explicitMaterialTransactionTonsMax,
	}, nil
}
