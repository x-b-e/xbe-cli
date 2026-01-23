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

type doEquipmentSuppliersCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	Name           string
	ContractNumber string
	BrokerID       string
}

func newDoEquipmentSuppliersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new equipment supplier",
		Long: `Create a new equipment supplier.

Required flags:
  --name    Supplier name
  --broker  Broker ID

Optional flags:
  --contract-number  Supplier contract number`,
		Example: `  # Create an equipment supplier
  xbe do equipment-suppliers create --name "Acme Equipment" --broker 123

  # Create with contract number
  xbe do equipment-suppliers create --name "Rental Co" --broker 123 --contract-number "EQ-2024-01"`,
		RunE: runDoEquipmentSuppliersCreate,
	}
	initDoEquipmentSuppliersCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentSuppliersCmd.AddCommand(newDoEquipmentSuppliersCreateCmd())
}

func initDoEquipmentSuppliersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Supplier name (required)")
	cmd.Flags().String("contract-number", "", "Supplier contract number")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("broker")
}

func runDoEquipmentSuppliersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentSuppliersCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.BrokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}
	if opts.ContractNumber != "" {
		attributes["contract-number"] = opts.ContractNumber
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "equipment-suppliers",
			"attributes": attributes,
			"relationships": map[string]any{
				"broker": map[string]any{
					"data": map[string]any{
						"type": "brokers",
						"id":   opts.BrokerID,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-suppliers", jsonBody)
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

	row := equipmentSupplierRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment supplier %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoEquipmentSuppliersCreateOptions(cmd *cobra.Command) (doEquipmentSuppliersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	contractNumber, _ := cmd.Flags().GetString("contract-number")
	brokerID, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentSuppliersCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		Name:           name,
		ContractNumber: contractNumber,
		BrokerID:       brokerID,
	}, nil
}

func equipmentSupplierRowFromSingle(resp jsonAPISingleResponse) equipmentSupplierRow {
	row := equipmentSupplierRow{
		ID:             resp.Data.ID,
		Name:           stringAttr(resp.Data.Attributes, "name"),
		ContractNumber: stringAttr(resp.Data.Attributes, "contract-number"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
