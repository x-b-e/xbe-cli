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

type doDriverDayAdjustmentPlansUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Content string
	StartAt string
}

func newDoDriverDayAdjustmentPlansUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a driver day adjustment plan",
		Long: `Update a driver day adjustment plan.

Optional:
  --content   Updated plan content
  --start-at  Updated start timestamp (ISO 8601)`,
		Example: `  # Update plan content
  xbe do driver-day-adjustment-plans update 123 --content "Adjusted for weather"

  # Update start timestamp
  xbe do driver-day-adjustment-plans update 123 --start-at "2025-01-16T06:00:00Z"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDriverDayAdjustmentPlansUpdate,
	}
	initDoDriverDayAdjustmentPlansUpdateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayAdjustmentPlansCmd.AddCommand(newDoDriverDayAdjustmentPlansUpdateCmd())
}

func initDoDriverDayAdjustmentPlansUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("content", "", "Updated plan content")
	cmd.Flags().String("start-at", "", "Updated start timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayAdjustmentPlansUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverDayAdjustmentPlansUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("content") {
		attributes["content"] = opts.Content
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "driver-day-adjustment-plans",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/driver-day-adjustment-plans/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := driverDayAdjustmentPlanRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated driver day adjustment plan %s\n", resp.Data.ID)
	return nil
}

func parseDoDriverDayAdjustmentPlansUpdateOptions(cmd *cobra.Command, args []string) (doDriverDayAdjustmentPlansUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	content, _ := cmd.Flags().GetString("content")
	startAt, _ := cmd.Flags().GetString("start-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayAdjustmentPlansUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Content: content,
		StartAt: startAt,
	}, nil
}
