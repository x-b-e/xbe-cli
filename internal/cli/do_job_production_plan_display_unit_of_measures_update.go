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

type doJobProductionPlanDisplayUnitOfMeasuresUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	ImportancePosition int
}

func newDoJobProductionPlanDisplayUnitOfMeasuresUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan display unit of measure",
		Long: `Update a job production plan display unit of measure.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The display unit of measure ID (required)

Flags:
  --importance-position  Update the importance position (0-based index)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update importance position
  xbe do job-production-plan-display-unit-of-measures update 123 --importance-position 1

  # Output as JSON
  xbe do job-production-plan-display-unit-of-measures update 123 --importance-position 2 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanDisplayUnitOfMeasuresUpdate,
	}
	initDoJobProductionPlanDisplayUnitOfMeasuresUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanDisplayUnitOfMeasuresCmd.AddCommand(newDoJobProductionPlanDisplayUnitOfMeasuresUpdateCmd())
}

func initDoJobProductionPlanDisplayUnitOfMeasuresUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("importance-position", 0, "Importance position (0-based index)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanDisplayUnitOfMeasuresUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanDisplayUnitOfMeasuresUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("importance-position") {
		attributes["importance-position"] = opts.ImportancePosition
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         opts.ID,
			"type":       "job-production-plan-display-unit-of-measures",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-display-unit-of-measures/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan display unit of measure %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanDisplayUnitOfMeasuresUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanDisplayUnitOfMeasuresUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	importancePosition, _ := cmd.Flags().GetInt("importance-position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanDisplayUnitOfMeasuresUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		ImportancePosition: importancePosition,
	}, nil
}
