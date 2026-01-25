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

type doTimeSheetApprovalsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	TimeSheet string
	Comment   string
}

func newDoTimeSheetApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve a time sheet",
		Long: `Approve a time sheet.

Time sheets must be in editing or submitted status, include a duration, and
may require a cost code allocation (for crew requirement time sheets).

Required flags:
  --time-sheet   Time sheet ID

Optional flags:
  --comment      Approval comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Approve a time sheet with a comment
  xbe do time-sheet-approvals create \
    --time-sheet 123 \
    --comment "Approved"

  # Approve a time sheet without a comment
  xbe do time-sheet-approvals create --time-sheet 123`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetApprovalsCreate,
	}
	initDoTimeSheetApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetApprovalsCmd.AddCommand(newDoTimeSheetApprovalsCreateCmd())
}

func initDoTimeSheetApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID")
	cmd.Flags().String("comment", "", "Approval comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetApprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetApprovalsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TimeSheet) == "" {
		err := fmt.Errorf("--time-sheet is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"time-sheet": map[string]any{
			"data": map[string]any{
				"type": "time-sheets",
				"id":   opts.TimeSheet,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-sheet-approvals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-approvals", jsonBody)
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

	row := buildTimeSheetApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet approval %s\n", row.ID)
	return nil
}

func parseDoTimeSheetApprovalsCreateOptions(cmd *cobra.Command) (doTimeSheetApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetApprovalsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		TimeSheet: timeSheet,
		Comment:   comment,
	}, nil
}
