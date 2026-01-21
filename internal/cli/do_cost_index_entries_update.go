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

type doCostIndexEntriesUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	StartOn string
	EndOn   string
	Value   string
}

func newDoCostIndexEntriesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing cost index entry",
		Long: `Update an existing cost index entry.

Provide the entry ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --start-on  Entry start date (format: YYYY-MM-DD)
  --end-on    Entry end date (format: YYYY-MM-DD)
  --value     Entry value`,
		Example: `  # Update value
  xbe do cost-index-entries update 123 --value 1.10

  # Update date range
  xbe do cost-index-entries update 123 --start-on "2024-01-01" --end-on "2024-06-30"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCostIndexEntriesUpdate,
	}
	initDoCostIndexEntriesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCostIndexEntriesCmd.AddCommand(newDoCostIndexEntriesUpdateCmd())
}

func initDoCostIndexEntriesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-on", "", "Entry start date")
	cmd.Flags().String("end-on", "", "Entry end date")
	cmd.Flags().String("value", "", "Entry value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostIndexEntriesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostIndexEntriesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("value") {
		attributes["value"] = opts.Value
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --start-on, --end-on, --value")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "cost-index-entries",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/cost-index-entries/"+opts.ID, jsonBody)
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

	row := buildCostIndexEntryRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated cost index entry %s\n", row.ID)
	return nil
}

func parseDoCostIndexEntriesUpdateOptions(cmd *cobra.Command, args []string) (doCostIndexEntriesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostIndexEntriesUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		StartOn: startOn,
		EndOn:   endOn,
		Value:   value,
	}, nil
}
