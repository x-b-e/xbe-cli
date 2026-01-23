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

type doMaterialTransactionTicketGeneratorsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	FormatRule       string
	OrganizationType string
	OrganizationID   string
}

func newDoMaterialTransactionTicketGeneratorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction ticket generator",
		Long: `Create a material transaction ticket generator.

Required flags:
  --format-rule         Ticket number format rule (required)
  --organization-type   Organization type (required)
  --organization-id     Organization ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a ticket generator for a broker
  xbe do material-transaction-ticket-generators create \
    --format-rule "MTX-{sequence}" \
    --organization-type brokers \
    --organization-id 123`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionTicketGeneratorsCreate,
	}
	initDoMaterialTransactionTicketGeneratorsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionTicketGeneratorsCmd.AddCommand(newDoMaterialTransactionTicketGeneratorsCreateCmd())
}

func initDoMaterialTransactionTicketGeneratorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("format-rule", "", "Ticket number format rule (required)")
	cmd.Flags().String("organization-type", "", "Organization type (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionTicketGeneratorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionTicketGeneratorsCreateOptions(cmd)
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

	if opts.FormatRule == "" {
		err := fmt.Errorf("--format-rule is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"format-rule": opts.FormatRule,
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transaction-ticket-generators",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-ticket-generators", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction ticket generator %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionTicketGeneratorsCreateOptions(cmd *cobra.Command) (doMaterialTransactionTicketGeneratorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	formatRule, _ := cmd.Flags().GetString("format-rule")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionTicketGeneratorsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		FormatRule:       formatRule,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}
