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

type doTimeSheetsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	StartAt             string
	EndAt               string
	BreakMinutes        int
	Notes               string
	SkipValidateOverlap bool
}

func newDoTimeSheetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time sheet",
		Long: `Update a time sheet.

Writable fields:
  --start-at              Start timestamp (ISO 8601)
  --end-at                End timestamp (ISO 8601)
  --break-minutes         Break minutes
  --notes                 Notes
  --skip-validate-overlap Skip overlap validation (true/false)

Arguments:
  <id>    Time sheet ID (required).`,
		Example: `  # Update a time sheet end time
  xbe do time-sheets update 123 --end-at 2026-01-01T17:00:00Z

  # Update notes and break minutes
  xbe do time-sheets update 123 --notes \"Updated\" --break-minutes 45`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetsUpdate,
	}
	initDoTimeSheetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetsCmd.AddCommand(newDoTimeSheetsUpdateCmd())
}

func initDoTimeSheetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().Int("break-minutes", 0, "Break minutes")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().Bool("skip-validate-overlap", false, "Skip overlap validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("break-minutes") {
		attributes["break-minutes"] = opts.BreakMinutes
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("skip-validate-overlap") {
		attributes["skip-validate-overlap"] = opts.SkipValidateOverlap
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "time-sheets",
		"id":         opts.ID,
		"attributes": attributes,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-sheets/"+opts.ID, jsonBody)
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

	row := buildTimeSheetRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time sheet %s\n", row.ID)
	return nil
}

func parseDoTimeSheetsUpdateOptions(cmd *cobra.Command, args []string) (doTimeSheetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	breakMinutes, _ := cmd.Flags().GetInt("break-minutes")
	notes, _ := cmd.Flags().GetString("notes")
	skipValidateOverlap, _ := cmd.Flags().GetBool("skip-validate-overlap")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doTimeSheetsUpdateOptions{}, fmt.Errorf("time sheet id is required")
	}

	return doTimeSheetsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  id,
		StartAt:             startAt,
		EndAt:               endAt,
		BreakMinutes:        breakMinutes,
		Notes:               notes,
		SkipValidateOverlap: skipValidateOverlap,
	}, nil
}
