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

type doProductionIncidentDetectorsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
	LookaheadOffset   int
	MinutesThreshold  int
	QuantityThreshold int
}

func newDoProductionIncidentDetectorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Run production incident detection",
		Long: `Run production incident detection for a job production plan.

Required flags:
  --job-production-plan   Job production plan ID (required)

Optional flags:
  --lookahead-offset      Lookahead offset in minutes (default: 60)
  --minutes-threshold     Minutes threshold in minutes (default: 60)
  --quantity-threshold    Quantity threshold in units (default: 75)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Run detection with defaults
  xbe do production-incident-detectors create --job-production-plan 123

  # Run detection with custom thresholds
  xbe do production-incident-detectors create \
    --job-production-plan 123 \
    --lookahead-offset 30 \
    --minutes-threshold 45 \
    --quantity-threshold 50

  # Output as JSON
  xbe do production-incident-detectors create --job-production-plan 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProductionIncidentDetectorsCreate,
	}
	initDoProductionIncidentDetectorsCreateFlags(cmd)
	return cmd
}

func init() {
	doProductionIncidentDetectorsCmd.AddCommand(newDoProductionIncidentDetectorsCreateCmd())
}

func initDoProductionIncidentDetectorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().Int("lookahead-offset", 0, "Lookahead offset in minutes")
	cmd.Flags().Int("minutes-threshold", 0, "Minutes threshold in minutes")
	cmd.Flags().Int("quantity-threshold", 0, "Quantity threshold in units")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProductionIncidentDetectorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProductionIncidentDetectorsCreateOptions(cmd)
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

	opts.JobProductionPlan = strings.TrimSpace(opts.JobProductionPlan)
	if opts.JobProductionPlan == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("lookahead-offset") {
		attributes["lookahead-offset"] = opts.LookaheadOffset
	}
	if cmd.Flags().Changed("minutes-threshold") {
		attributes["minutes-threshold"] = opts.MinutesThreshold
	}
	if cmd.Flags().Changed("quantity-threshold") {
		attributes["quantity-threshold"] = opts.QuantityThreshold
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "production-incident-detectors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/production-incident-detectors", jsonBody)
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

	details := buildProductionIncidentDetectorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	count := 0
	if details.Incidents != nil {
		count = countConstraintItems(details.Incidents)
	}
	if count > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Created production incident detector %s (%d incidents)\n", details.ID, count)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created production incident detector %s\n", details.ID)
	return nil
}

func parseDoProductionIncidentDetectorsCreateOptions(cmd *cobra.Command) (doProductionIncidentDetectorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	lookaheadOffset, _ := cmd.Flags().GetInt("lookahead-offset")
	minutesThreshold, _ := cmd.Flags().GetInt("minutes-threshold")
	quantityThreshold, _ := cmd.Flags().GetInt("quantity-threshold")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProductionIncidentDetectorsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
		LookaheadOffset:   lookaheadOffset,
		MinutesThreshold:  minutesThreshold,
		QuantityThreshold: quantityThreshold,
	}, nil
}
