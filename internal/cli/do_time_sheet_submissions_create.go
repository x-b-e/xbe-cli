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

type doTimeSheetSubmissionsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	TimeSheetID string
	Comment     string
}

type timeSheetSubmissionRow struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newDoTimeSheetSubmissionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Submit a time sheet",
		Long: `Submit a time sheet.

The time sheet must be in editing or rejected status and include a duration.

Required:
  --time-sheet  Time sheet ID (required)

Optional:
  --comment     Status change comment`,
		Example: `  # Submit a time sheet
  xbe do time-sheet-submissions create --time-sheet 123

  # Submit with a comment
  xbe do time-sheet-submissions create --time-sheet 123 --comment "Ready for approval"

  # Output as JSON
  xbe do time-sheet-submissions create --time-sheet 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetSubmissionsCreate,
	}
	initDoTimeSheetSubmissionsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetSubmissionsCmd.AddCommand(newDoTimeSheetSubmissionsCreateCmd())
}

func initDoTimeSheetSubmissionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID (required)")
	cmd.Flags().String("comment", "", "Status change comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("time-sheet")
}

func runDoTimeSheetSubmissionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetSubmissionsCreateOptions(cmd)
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

	if opts.TimeSheetID == "" {
		err := fmt.Errorf("--time-sheet is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"time-sheet": map[string]any{
			"data": map[string]any{
				"type": "time-sheets",
				"id":   opts.TimeSheetID,
			},
		},
	}

	data := map[string]any{
		"type":          "time-sheet-submissions",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-submissions", jsonBody)
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

	row := buildTimeSheetSubmissionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet submission %s\n", row.ID)
	return nil
}

func parseDoTimeSheetSubmissionsCreateOptions(cmd *cobra.Command) (doTimeSheetSubmissionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheetID, _ := cmd.Flags().GetString("time-sheet")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetSubmissionsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		TimeSheetID: timeSheetID,
		Comment:     comment,
	}, nil
}

func buildTimeSheetSubmissionRowFromSingle(resp jsonAPISingleResponse) timeSheetSubmissionRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeSheetSubmissionRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}
	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		row.TimeSheetID = rel.Data.ID
	}
	return row
}
