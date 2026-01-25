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

type doEquipmentSuppliersUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	Name           string
	ContractNumber string
}

func newDoEquipmentSuppliersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment supplier",
		Long: `Update an equipment supplier.

Optional flags:
  --name             Supplier name
  --contract-number  Supplier contract number`,
		Example: `  # Update supplier name
  xbe do equipment-suppliers update 123 --name "New Name"

  # Update contract number
  xbe do equipment-suppliers update 123 --contract-number "EQ-2024-02"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentSuppliersUpdate,
	}
	initDoEquipmentSuppliersUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentSuppliersCmd.AddCommand(newDoEquipmentSuppliersUpdateCmd())
}

func initDoEquipmentSuppliersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Supplier name")
	cmd.Flags().String("contract-number", "", "Supplier contract number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentSuppliersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentSuppliersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("contract-number") {
		attributes["contract-number"] = opts.ContractNumber
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "equipment-suppliers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-suppliers/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment supplier %s\n", row.ID)
	return nil
}

func parseDoEquipmentSuppliersUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentSuppliersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	contractNumber, _ := cmd.Flags().GetString("contract-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentSuppliersUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		Name:           name,
		ContractNumber: contractNumber,
	}, nil
}
