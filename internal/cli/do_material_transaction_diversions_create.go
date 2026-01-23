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

type doMaterialTransactionDiversionsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	MaterialTransaction  string
	NewJobSite           string
	NewDeliveryDate      string
	DivertedTonsExplicit string
	DriverInstructions   string
}

func newDoMaterialTransactionDiversionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction diversion",
		Long: `Create a material transaction diversion.

Required flags:
  --material-transaction   Material transaction ID

Optional flags:
  --new-job-site            New job site ID
  --new-delivery-date       New delivery date (YYYY-MM-DD)
  --diverted-tons-explicit  Explicit diverted tons (must be > 0 and <= transaction tons)
  --driver-instructions     Driver instructions

Notes:
  - new-delivery-date must be on or after the material transaction date.
  - material-transaction can only be diverted once.`,
		Example: `  # Create a diversion with a new delivery date
  xbe do material-transaction-diversions create --material-transaction 123 --new-delivery-date 2025-01-02

  # Create a diversion with driver instructions
  xbe do material-transaction-diversions create --material-transaction 123 --driver-instructions "Call dispatch"`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionDiversionsCreate,
	}
	initDoMaterialTransactionDiversionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionDiversionsCmd.AddCommand(newDoMaterialTransactionDiversionsCreateCmd())
}

func initDoMaterialTransactionDiversionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-transaction", "", "Material transaction ID (required)")
	cmd.Flags().String("new-job-site", "", "New job site ID")
	cmd.Flags().String("new-delivery-date", "", "New delivery date (YYYY-MM-DD)")
	cmd.Flags().String("diverted-tons-explicit", "", "Explicit diverted tons")
	cmd.Flags().String("driver-instructions", "", "Driver instructions")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionDiversionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionDiversionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaterialTransaction) == "" {
		err := fmt.Errorf("--material-transaction is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.NewDeliveryDate != "" {
		attributes["new-delivery-date"] = opts.NewDeliveryDate
	}
	if opts.DivertedTonsExplicit != "" {
		attributes["diverted-tons-explicit"] = opts.DivertedTonsExplicit
	}
	if opts.DriverInstructions != "" {
		attributes["driver-instructions"] = opts.DriverInstructions
	}

	relationships := map[string]any{
		"material-transaction": map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		},
	}

	if opts.NewJobSite != "" {
		relationships["new-job-site"] = map[string]any{
			"data": map[string]any{
				"type": "job-sites",
				"id":   opts.NewJobSite,
			},
		}
	}

	data := map[string]any{
		"type":          "material-transaction-diversions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-diversions", jsonBody)
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

	row := materialTransactionDiversionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction diversion %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionDiversionsCreateOptions(cmd *cobra.Command) (doMaterialTransactionDiversionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	newJobSite, _ := cmd.Flags().GetString("new-job-site")
	newDeliveryDate, _ := cmd.Flags().GetString("new-delivery-date")
	divertedTonsExplicit, _ := cmd.Flags().GetString("diverted-tons-explicit")
	driverInstructions, _ := cmd.Flags().GetString("driver-instructions")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionDiversionsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		MaterialTransaction:  materialTransaction,
		NewJobSite:           newJobSite,
		NewDeliveryDate:      newDeliveryDate,
		DivertedTonsExplicit: divertedTonsExplicit,
		DriverInstructions:   driverInstructions,
	}, nil
}
