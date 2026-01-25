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

type doTimeSheetNoShowsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	TimeSheet    string
	NoShowReason string
}

func newDoTimeSheetNoShowsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time sheet no-show",
		Long: `Create a time sheet no-show.

Required flags:
  --time-sheet      Time sheet ID
  --no-show-reason  Reason for the no-show

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a no-show
  xbe do time-sheet-no-shows create \
    --time-sheet 123 \
    --no-show-reason "No show"`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetNoShowsCreate,
	}
	initDoTimeSheetNoShowsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetNoShowsCmd.AddCommand(newDoTimeSheetNoShowsCreateCmd())
}

func initDoTimeSheetNoShowsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-sheet", "", "Time sheet ID")
	cmd.Flags().String("no-show-reason", "", "Reason for the no-show")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetNoShowsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetNoShowsCreateOptions(cmd)
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
	if strings.TrimSpace(opts.NoShowReason) == "" {
		err := fmt.Errorf("--no-show-reason is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"no-show-reason": opts.NoShowReason,
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
			"type":          "time-sheet-no-shows",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-no-shows", jsonBody)
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

	row := buildTimeSheetNoShowRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet no-show %s\n", row.ID)
	return nil
}

func parseDoTimeSheetNoShowsCreateOptions(cmd *cobra.Command) (doTimeSheetNoShowsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	noShowReason, _ := cmd.Flags().GetString("no-show-reason")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetNoShowsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		TimeSheet:    timeSheet,
		NoShowReason: noShowReason,
	}, nil
}
