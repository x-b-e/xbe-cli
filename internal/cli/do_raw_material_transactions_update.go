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

type doRawMaterialTransactionsUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	TicketJobNumber string
}

type rawMaterialTransactionUpdateResult struct {
	ID              string `json:"id"`
	TicketJobNumber string `json:"ticket_job_number,omitempty"`
}

func newDoRawMaterialTransactionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a raw material transaction",
		Long: `Update a raw material transaction.

Only the fields you specify will be updated. Raw material transactions are
admin-only updates.

Fields:
  --ticket-job-number    Ticket job number

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update the ticket job number
  xbe do raw-material-transactions update 123 --ticket-job-number "JOB-001"

  # JSON output
  xbe do raw-material-transactions update 123 --ticket-job-number "JOB-001" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRawMaterialTransactionsUpdate,
	}
	initDoRawMaterialTransactionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRawMaterialTransactionsCmd.AddCommand(newDoRawMaterialTransactionsUpdateCmd())
}

func initDoRawMaterialTransactionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("ticket-job-number", "", "Ticket job number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRawMaterialTransactionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRawMaterialTransactionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("ticket-job-number") {
		attributes["ticket-job-number"] = opts.TicketJobNumber
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "raw-material-transactions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/raw-material-transactions/"+opts.ID, jsonBody)
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

	if opts.JSON {
		result := rawMaterialTransactionUpdateResult{
			ID:              resp.Data.ID,
			TicketJobNumber: stringAttr(resp.Data.Attributes, "ticket-job-number"),
		}
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated raw material transaction %s\n", resp.Data.ID)
	return nil
}

func parseDoRawMaterialTransactionsUpdateOptions(cmd *cobra.Command, args []string) (doRawMaterialTransactionsUpdateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doRawMaterialTransactionsUpdateOptions{}, err
	}
	ticketJobNumber, err := cmd.Flags().GetString("ticket-job-number")
	if err != nil {
		return doRawMaterialTransactionsUpdateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doRawMaterialTransactionsUpdateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doRawMaterialTransactionsUpdateOptions{}, err
	}

	return doRawMaterialTransactionsUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		TicketJobNumber: ticketJobNumber,
	}, nil
}
