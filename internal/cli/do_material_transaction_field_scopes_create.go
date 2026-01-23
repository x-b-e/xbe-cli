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

type doMaterialTransactionFieldScopesCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TenderJobScheduleShift string
}

func newDoMaterialTransactionFieldScopesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction field scope",
		Long: `Create a material transaction field scope.

Field scopes are computed from a tender job schedule shift to help match
material transactions to jobs, sites, and material types.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a field scope from a tender job schedule shift
  xbe do material-transaction-field-scopes create \
    --tender-job-schedule-shift 123

  # Output as JSON
  xbe do material-transaction-field-scopes create \
    --tender-job-schedule-shift 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionFieldScopesCreate,
	}
	initDoMaterialTransactionFieldScopesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionFieldScopesCmd.AddCommand(newDoMaterialTransactionFieldScopesCreateCmd())
}

func initDoMaterialTransactionFieldScopesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("tender-job-schedule-shift")
}

func runDoMaterialTransactionFieldScopesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionFieldScopesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TenderJobScheduleShift) == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"tender-job-schedule-shifts": map[string]any{
			"data": []any{
				map[string]any{
					"type": "tender-job-schedule-shifts",
					"id":   opts.TenderJobScheduleShift,
				},
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transaction-field-scopes",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-field-scopes", jsonBody)
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

	row := materialTransactionFieldScopeRow{
		ID:            resp.Data.ID,
		TicketNumber:  stringAttr(resp.Data.Attributes, "ticket-number"),
		TransactionAt: formatDateTime(stringAttr(resp.Data.Attributes, "transaction-at")),
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction field scope %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionFieldScopesCreateOptions(cmd *cobra.Command) (doMaterialTransactionFieldScopesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionFieldScopesCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}
