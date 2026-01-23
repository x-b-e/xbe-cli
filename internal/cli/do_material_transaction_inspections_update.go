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

type doMaterialTransactionInspectionsUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Note     string
	Status   string
	Strategy string
}

func newDoMaterialTransactionInspectionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material transaction inspection",
		Long: `Update a material transaction inspection.

Optional attributes:
  --status     Inspection status (open,closed)
  --strategy   Inspection strategy (delivery_site_personnel)
  --note       Inspection note`,
		Example: `  # Update inspection note and status
  xbe do material-transaction-inspections update 123 \\
    --note "Closing inspection" \\
    --status closed`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTransactionInspectionsUpdate,
	}
	initDoMaterialTransactionInspectionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionInspectionsCmd.AddCommand(newDoMaterialTransactionInspectionsUpdateCmd())
}

func initDoMaterialTransactionInspectionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Inspection status (open,closed)")
	cmd.Flags().String("strategy", "", "Inspection strategy (delivery_site_personnel)")
	cmd.Flags().String("note", "", "Inspection note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionInspectionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionInspectionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("strategy") {
		attributes["strategy"] = opts.Strategy
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-transaction-inspections",
		"id":   opts.ID,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-transaction-inspections/"+opts.ID, jsonBody)
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

	row := materialTransactionInspectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material transaction inspection %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionInspectionsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTransactionInspectionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	strategy, _ := cmd.Flags().GetString("strategy")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionInspectionsUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		Note:     note,
		Status:   status,
		Strategy: strategy,
	}, nil
}
