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

type doMaterialTransactionTicketGeneratorsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	FormatRule string
}

func newDoMaterialTransactionTicketGeneratorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material transaction ticket generator",
		Long: `Update a material transaction ticket generator.

Optional flags:
  --format-rule    Ticket number format rule

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a ticket generator format rule
  xbe do material-transaction-ticket-generators update 123 --format-rule "MTX-{sequence}-A"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTransactionTicketGeneratorsUpdate,
	}
	initDoMaterialTransactionTicketGeneratorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionTicketGeneratorsCmd.AddCommand(newDoMaterialTransactionTicketGeneratorsUpdateCmd())
}

func initDoMaterialTransactionTicketGeneratorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("format-rule", "", "Ticket number format rule")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionTicketGeneratorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionTicketGeneratorsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("format-rule") {
		attributes["format-rule"] = opts.FormatRule
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "material-transaction-ticket-generators",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-transaction-ticket-generators/"+opts.ID, jsonBody)
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

	row := buildMaterialTransactionTicketGeneratorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material transaction ticket generator %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionTicketGeneratorsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTransactionTicketGeneratorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	formatRule, _ := cmd.Flags().GetString("format-rule")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionTicketGeneratorsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		FormatRule: formatRule,
	}, nil
}
