package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanTrailerClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlanTrailerClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan trailer classification details",
		Long: `Show the full details of a job production plan trailer classification.

Output Fields:
  ID
  Job Production Plan ID
  Trailer Classification ID
  Trailer Classification Equivalent IDs
  Gross Weight Legal Limit (explicit)
  Gross Weight Legal Limit
  Explicit Material Transaction Tons Max
  Material Transaction Tons Max

Arguments:
  <id>    The job production plan trailer classification ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job production plan trailer classification
  xbe view job-production-plan-trailer-classifications show 123

  # Output as JSON
  xbe view job-production-plan-trailer-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanTrailerClassificationsShow,
	}
	initJobProductionPlanTrailerClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanTrailerClassificationsCmd.AddCommand(newJobProductionPlanTrailerClassificationsShowCmd())
}

func initJobProductionPlanTrailerClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanTrailerClassificationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanTrailerClassificationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan trailer classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-trailer-classifications]", "trailer-classification-equivalent-ids,gross-weight-legal-limit-lbs-explicit,gross-weight-legal-limit-lbs,explicit-material-transaction-tons-max,material-transaction-tons-max,job-production-plan,trailer-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-trailer-classifications/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobProductionPlanTrailerClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanTrailerClassificationDetails(cmd, details)
}

func parseJobProductionPlanTrailerClassificationsShowOptions(cmd *cobra.Command) (jobProductionPlanTrailerClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanTrailerClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobProductionPlanTrailerClassificationDetails(cmd *cobra.Command, details jobProductionPlanTrailerClassificationRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlan)
	}
	if details.TrailerClassification != "" {
		fmt.Fprintf(out, "Trailer Classification ID: %s\n", details.TrailerClassification)
	}
	if len(details.TrailerClassificationEquivalentIDs) > 0 {
		fmt.Fprintf(out, "Trailer Classification Equivalent IDs: %s\n", strings.Join(details.TrailerClassificationEquivalentIDs, ", "))
	}
	if details.GrossWeightLegalLimitLbsExplicit > 0 {
		fmt.Fprintf(out, "Gross Weight Legal Limit (explicit): %s\n", formatOptionalFloat(details.GrossWeightLegalLimitLbsExplicit))
	}
	if details.GrossWeightLegalLimitLbs > 0 {
		fmt.Fprintf(out, "Gross Weight Legal Limit: %s\n", formatOptionalFloat(details.GrossWeightLegalLimitLbs))
	}
	if details.ExplicitMaterialTransactionTonsMax > 0 {
		fmt.Fprintf(out, "Explicit Material Transaction Tons Max: %s\n", formatOptionalFloat(details.ExplicitMaterialTransactionTonsMax))
	}
	if details.MaterialTransactionTonsMax > 0 {
		fmt.Fprintf(out, "Material Transaction Tons Max: %s\n", formatOptionalFloat(details.MaterialTransactionTonsMax))
	}

	return nil
}
