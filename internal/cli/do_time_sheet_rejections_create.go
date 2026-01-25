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

type doTimeSheetRejectionsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	TimeSheet string
	Comment   string
}

func newDoTimeSheetRejectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Reject a time sheet",
		Long: `Reject a submitted time sheet.

Required flags:
  --time-sheet   Time sheet ID (required)

Optional flags:
  --comment      Comment explaining the rejection`,
		Example: `  # Reject a time sheet
  xbe do time-sheet-rejections create --time-sheet 123

  # Reject with a comment
  xbe do time-sheet-rejections create --time-sheet 123 --comment "Missing backup"

  # JSON output
  xbe do time-sheet-rejections create --time-sheet 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetRejectionsCreate,
	}
	initDoTimeSheetRejectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetRejectionsCmd.AddCommand(newDoTimeSheetRejectionsCreateCmd())
}

func initDoTimeSheetRejectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID (required)")
	cmd.Flags().String("comment", "", "Comment explaining the rejection")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetRejectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetRejectionsCreateOptions(cmd)
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
	if strings.TrimSpace(opts.Comment) != "" {
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

	data := map[string]any{
		"type":          "time-sheet-rejections",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-rejections", jsonBody)
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

	row := buildTimeSheetRejectionRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet rejection %s\n", row.ID)
	return nil
}

func parseDoTimeSheetRejectionsCreateOptions(cmd *cobra.Command) (doTimeSheetRejectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetRejectionsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		TimeSheet: timeSheet,
		Comment:   comment,
	}, nil
}
