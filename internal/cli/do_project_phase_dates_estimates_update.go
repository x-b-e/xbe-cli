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

type doProjectPhaseDatesEstimatesUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	CreatedBy string
	StartDate string
	EndDate   string
}

func newDoProjectPhaseDatesEstimatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project phase dates estimate",
		Long: `Update a project phase dates estimate.

All flags are optional. Only provided flags will be updated.

Attributes:
  --start-date  Estimated start date (YYYY-MM-DD)
  --end-date    Estimated end date (YYYY-MM-DD)

Relationships:
  --created-by  Creator user ID`,
		Example: `  # Update the end date
  xbe do project-phase-dates-estimates update 123 --end-date 2025-02-01

  # Update the start date
  xbe do project-phase-dates-estimates update 123 --start-date 2025-01-05`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseDatesEstimatesUpdate,
	}
	initDoProjectPhaseDatesEstimatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseDatesEstimatesCmd.AddCommand(newDoProjectPhaseDatesEstimatesUpdateCmd())
}

func initDoProjectPhaseDatesEstimatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-date", "", "Estimated start date (YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "Estimated end date (YYYY-MM-DD)")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseDatesEstimatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseDatesEstimatesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("start-date") {
		if strings.TrimSpace(opts.StartDate) == "" {
			err := fmt.Errorf("--start-date cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["start-date"] = opts.StartDate
	}
	if cmd.Flags().Changed("end-date") {
		if strings.TrimSpace(opts.EndDate) == "" {
			err := fmt.Errorf("--end-date cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["end-date"] = opts.EndDate
	}
	if cmd.Flags().Changed("created-by") {
		if strings.TrimSpace(opts.CreatedBy) == "" {
			err := fmt.Errorf("--created-by cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-phase-dates-estimates",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-phase-dates-estimates/"+opts.ID, jsonBody)
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

	row := projectPhaseDatesEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project phase dates estimate %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseDatesEstimatesUpdateOptions(cmd *cobra.Command, args []string) (doProjectPhaseDatesEstimatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseDatesEstimatesUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		CreatedBy: createdBy,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}
