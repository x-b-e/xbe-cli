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

type doJobProductionPlanTrailerClassificationsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	JobProductionPlan                  string
	TrailerClassification              string
	TrailerClassificationEquivalentIDs []string
	GrossWeightLegalLimitLbsExplicit   string
	ExplicitMaterialTransactionTonsMax string
}

func newDoJobProductionPlanTrailerClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add a trailer classification to a job production plan",
		Long: `Add a trailer classification to a job production plan.

Required flags:
  --job-production-plan      Job production plan ID
  --trailer-classification   Trailer classification ID

Optional flags:
  --trailer-classification-equivalent-ids  Equivalent trailer classification IDs (comma-separated or repeated)
  --gross-weight-legal-limit-lbs-explicit  Explicit gross weight legal limit (lbs)
  --explicit-material-transaction-tons-max Explicit material transaction tons max

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Add a trailer classification to a job production plan
  xbe do job-production-plan-trailer-classifications create \
    --job-production-plan 123 \
    --trailer-classification 456

  # Add with explicit weight and tons max
  xbe do job-production-plan-trailer-classifications create \
    --job-production-plan 123 \
    --trailer-classification 456 \
    --gross-weight-legal-limit-lbs-explicit 80000 \
    --explicit-material-transaction-tons-max 20

  # Output as JSON
  xbe do job-production-plan-trailer-classifications create \
    --job-production-plan 123 \
    --trailer-classification 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanTrailerClassificationsCreate,
	}
	initDoJobProductionPlanTrailerClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanTrailerClassificationsCmd.AddCommand(newDoJobProductionPlanTrailerClassificationsCreateCmd())
}

func initDoJobProductionPlanTrailerClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().StringSlice("trailer-classification-equivalent-ids", nil, "Equivalent trailer classification IDs (comma-separated or repeated)")
	cmd.Flags().String("gross-weight-legal-limit-lbs-explicit", "", "Explicit gross weight legal limit (lbs)")
	cmd.Flags().String("explicit-material-transaction-tons-max", "", "Explicit material transaction tons max")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("trailer-classification")
}

func runDoJobProductionPlanTrailerClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanTrailerClassificationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.JobProductionPlan) == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.TrailerClassification) == "" {
		err := fmt.Errorf("--trailer-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("trailer-classification-equivalent-ids") {
		attributes["trailer-classification-equivalent-ids"] = opts.TrailerClassificationEquivalentIDs
	}
	if cmd.Flags().Changed("gross-weight-legal-limit-lbs-explicit") {
		attributes["gross-weight-legal-limit-lbs-explicit"] = opts.GrossWeightLegalLimitLbsExplicit
	}
	if cmd.Flags().Changed("explicit-material-transaction-tons-max") {
		attributes["explicit-material-transaction-tons-max"] = opts.ExplicitMaterialTransactionTonsMax
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"trailer-classification": map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-trailer-classifications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-trailer-classifications", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan trailer classification %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanTrailerClassificationsCreateOptions(cmd *cobra.Command) (doJobProductionPlanTrailerClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	trailerClassificationEquivalentIDs, _ := cmd.Flags().GetStringSlice("trailer-classification-equivalent-ids")
	grossWeightLegalLimitLbsExplicit, _ := cmd.Flags().GetString("gross-weight-legal-limit-lbs-explicit")
	explicitMaterialTransactionTonsMax, _ := cmd.Flags().GetString("explicit-material-transaction-tons-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanTrailerClassificationsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		JobProductionPlan:                  jobProductionPlan,
		TrailerClassification:              trailerClassification,
		TrailerClassificationEquivalentIDs: trailerClassificationEquivalentIDs,
		GrossWeightLegalLimitLbsExplicit:   grossWeightLegalLimitLbsExplicit,
		ExplicitMaterialTransactionTonsMax: explicitMaterialTransactionTonsMax,
	}, nil
}
