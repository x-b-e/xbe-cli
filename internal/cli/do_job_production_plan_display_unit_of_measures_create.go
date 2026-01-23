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

type doJobProductionPlanDisplayUnitOfMeasuresCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	JobProductionPlan  string
	UnitOfMeasure      string
	ImportancePosition int
}

func newDoJobProductionPlanDisplayUnitOfMeasuresCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add a display unit of measure to a job production plan",
		Long: `Add a display unit of measure to a job production plan.

Units of measure must be unique per job production plan and must use a metric
of area, mass, or volume.

Required flags:
  --job-production-plan  Job production plan ID
  --unit-of-measure      Unit of measure ID

Optional flags:
  --importance-position  Importance position (0-based index)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Add a unit of measure to a job production plan
  xbe do job-production-plan-display-unit-of-measures create \
    --job-production-plan 123 \
    --unit-of-measure 456

  # Add with explicit importance position
  xbe do job-production-plan-display-unit-of-measures create \
    --job-production-plan 123 \
    --unit-of-measure 456 \
    --importance-position 0

  # Output as JSON
  xbe do job-production-plan-display-unit-of-measures create \
    --job-production-plan 123 \
    --unit-of-measure 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanDisplayUnitOfMeasuresCreate,
	}
	initDoJobProductionPlanDisplayUnitOfMeasuresCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanDisplayUnitOfMeasuresCmd.AddCommand(newDoJobProductionPlanDisplayUnitOfMeasuresCreateCmd())
}

func initDoJobProductionPlanDisplayUnitOfMeasuresCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().Int("importance-position", 0, "Importance position (0-based index)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("unit-of-measure")
}

func runDoJobProductionPlanDisplayUnitOfMeasuresCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanDisplayUnitOfMeasuresCreateOptions(cmd)
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

	if strings.TrimSpace(opts.UnitOfMeasure) == "" {
		err := fmt.Errorf("--unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("importance-position") {
		attributes["importance-position"] = opts.ImportancePosition
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-display-unit-of-measures",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-display-unit-of-measures", jsonBody)
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

	row := buildJobProductionPlanDisplayUnitOfMeasureRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan display unit of measure %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanDisplayUnitOfMeasuresCreateOptions(cmd *cobra.Command) (doJobProductionPlanDisplayUnitOfMeasuresCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	importancePosition, _ := cmd.Flags().GetInt("importance-position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanDisplayUnitOfMeasuresCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		JobProductionPlan:  jobProductionPlan,
		UnitOfMeasure:      unitOfMeasure,
		ImportancePosition: importancePosition,
	}, nil
}
