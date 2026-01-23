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

type doInvoiceGenerationsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	OrganizationType string
	OrganizationID   string
	TimeCardIDs      []string
	Note             string
}

func newDoInvoiceGenerationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an invoice generation",
		Long: `Create an invoice generation.

Required flags:
  --organization-type  Organization type (brokers, customers, truckers) (required)
  --organization-id    Organization ID (required)

Optional flags:
  --time-card-ids       Time card IDs (comma-separated or repeated)
  --note                Generation note`,
		Example: `  # Create an invoice generation for specific time cards
  xbe do invoice-generations create \
    --organization-type brokers \
    --organization-id 123 \
    --time-card-ids 456,789 \
    --note "End of week run"

  # Output as JSON
  xbe do invoice-generations create \
    --organization-type brokers \
    --organization-id 123 \
    --time-card-ids 456,789 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoInvoiceGenerationsCreate,
	}
	initDoInvoiceGenerationsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceGenerationsCmd.AddCommand(newDoInvoiceGenerationsCreateCmd())
}

func initDoInvoiceGenerationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-type", "", "Organization type (brokers, customers, truckers) (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().StringSlice("time-card-ids", nil, "Time card IDs (comma-separated or repeated)")
	cmd.Flags().String("note", "", "Generation note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInvoiceGenerationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceGenerationsCreateOptions(cmd)
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

	opts.OrganizationType = strings.TrimSpace(opts.OrganizationType)
	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	opts.OrganizationID = strings.TrimSpace(opts.OrganizationID)
	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	timeCardIDs := make([]string, 0, len(opts.TimeCardIDs))
	for _, id := range opts.TimeCardIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			timeCardIDs = append(timeCardIDs, trimmed)
		}
	}

	attributes := map[string]any{}
	if len(timeCardIDs) > 0 {
		attributes["time-card-ids"] = timeCardIDs
	}
	if strings.TrimSpace(opts.Note) != "" {
		attributes["note"] = opts.Note
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
			"type":          "invoice-generations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-generations", jsonBody)
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

	row := invoiceGenerationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice generation %s\n", row.ID)
	return nil
}

func parseDoInvoiceGenerationsCreateOptions(cmd *cobra.Command) (doInvoiceGenerationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	timeCardIDs, _ := cmd.Flags().GetStringSlice("time-card-ids")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceGenerationsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
		TimeCardIDs:      timeCardIDs,
		Note:             note,
	}, nil
}
