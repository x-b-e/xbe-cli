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

type doTimeSheetNoShowsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	NoShowReason string
}

func newDoTimeSheetNoShowsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time sheet no-show",
		Long: `Update a time sheet no-show.

Provide the no-show ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --no-show-reason  Reason for the no-show`,
		Example: `  # Update a no-show reason
  xbe do time-sheet-no-shows update 123 --no-show-reason "Updated reason"

  # Output as JSON
  xbe do time-sheet-no-shows update 123 --no-show-reason "Updated reason" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetNoShowsUpdate,
	}
	initDoTimeSheetNoShowsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetNoShowsCmd.AddCommand(newDoTimeSheetNoShowsUpdateCmd())
}

func initDoTimeSheetNoShowsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("no-show-reason", "", "Reason for the no-show")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetNoShowsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetNoShowsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("no-show-reason") {
		attributes["no-show-reason"] = opts.NoShowReason
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --no-show-reason")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-sheet-no-shows",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/time-sheet-no-shows/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time sheet no-show %s\n", row.ID)
	return nil
}

func parseDoTimeSheetNoShowsUpdateOptions(cmd *cobra.Command, args []string) (doTimeSheetNoShowsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noShowReason, _ := cmd.Flags().GetString("no-show-reason")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetNoShowsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		NoShowReason: noShowReason,
	}, nil
}
