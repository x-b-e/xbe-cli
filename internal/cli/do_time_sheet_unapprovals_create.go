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

type doTimeSheetUnapprovalsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	TimeSheet string
	Comment   string
}

func newDoTimeSheetUnapprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unapprove a time sheet",
		Long: `Unapprove a time sheet.

Time sheets must be approved to be unapproved, returning them to submitted status.

Required flags:
  --time-sheet   Time sheet ID

Optional flags:
  --comment      Unapproval comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Unapprove a time sheet with a comment
  xbe do time-sheet-unapprovals create \
    --time-sheet 123 \
    --comment "Needs review"

  # Unapprove a time sheet without a comment
  xbe do time-sheet-unapprovals create --time-sheet 123`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetUnapprovalsCreate,
	}
	initDoTimeSheetUnapprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetUnapprovalsCmd.AddCommand(newDoTimeSheetUnapprovalsCreateCmd())
}

func initDoTimeSheetUnapprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID")
	cmd.Flags().String("comment", "", "Unapproval comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetUnapprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetUnapprovalsCreateOptions(cmd)
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
			"type":          "time-sheet-unapprovals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-unapprovals", jsonBody)
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

	row := buildTimeSheetUnapprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet unapproval %s\n", row.ID)
	return nil
}

func parseDoTimeSheetUnapprovalsCreateOptions(cmd *cobra.Command) (doTimeSheetUnapprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetUnapprovalsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		TimeSheet: timeSheet,
		Comment:   comment,
	}, nil
}
