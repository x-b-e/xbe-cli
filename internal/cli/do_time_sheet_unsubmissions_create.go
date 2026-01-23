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

type doTimeSheetUnsubmissionsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	TimeSheetID string
	Comment     string
}

func newDoTimeSheetUnsubmissionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unsubmit a time sheet",
		Long: `Unsubmit a time sheet.

The time sheet must currently be in submitted status.

Required:
  --time-sheet  Time sheet ID (required)

Optional:
  --comment     Status change comment`,
		Example: `  # Unsubmit a time sheet
  xbe do time-sheet-unsubmissions create --time-sheet 123

  # Unsubmit with a comment
  xbe do time-sheet-unsubmissions create --time-sheet 123 --comment "Needs edits"

  # Output as JSON
  xbe do time-sheet-unsubmissions create --time-sheet 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetUnsubmissionsCreate,
	}
	initDoTimeSheetUnsubmissionsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetUnsubmissionsCmd.AddCommand(newDoTimeSheetUnsubmissionsCreateCmd())
}

func initDoTimeSheetUnsubmissionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID (required)")
	cmd.Flags().String("comment", "", "Status change comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("time-sheet")
}

func runDoTimeSheetUnsubmissionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetUnsubmissionsCreateOptions(cmd)
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
		"type":          "time-sheet-unsubmissions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-unsubmissions", jsonBody)
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

	row := buildTimeSheetUnsubmissionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet unsubmission %s\n", row.ID)
	return nil
}

func parseDoTimeSheetUnsubmissionsCreateOptions(cmd *cobra.Command) (doTimeSheetUnsubmissionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheetID, _ := cmd.Flags().GetString("time-sheet")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetUnsubmissionsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		TimeSheetID: timeSheetID,
		Comment:     comment,
	}, nil
}
