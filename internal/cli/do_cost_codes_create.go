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

type doCostCodesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Code        string
	Description string
	IsActive    bool
	CustomerID  string
	TruckerID   string
}

func newDoCostCodesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cost code",
		Long: `Create a new cost code.

Cost codes must be associated with either a customer or a trucker.

Required flags:
  --code    Cost code value

Association (one required):
  --customer        Customer ID
  --trucker         Trucker ID

Optional flags:
  --description     Description of the cost code
  --active          Set as active (default: true)`,
		Example: `  # Create a cost code for a customer
  xbe do cost-codes create --code "MAT-001" --description "Materials" --customer 123

  # Create a cost code for a trucker
  xbe do cost-codes create --code "FUEL-001" --description "Fuel costs" --trucker 456`,
		RunE: runDoCostCodesCreate,
	}
	initDoCostCodesCreateFlags(cmd)
	return cmd
}

func init() {
	doCostCodesCmd.AddCommand(newDoCostCodesCreateCmd())
}

func initDoCostCodesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Cost code value (required)")
	cmd.Flags().String("description", "", "Description of the cost code")
	cmd.Flags().Bool("active", true, "Set as active")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("code")
}

func runDoCostCodesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCostCodesCreateOptions(cmd)
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

	attributes := map[string]any{
		"code":      opts.Code,
		"is-active": opts.IsActive,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	relationships := map[string]any{}

	if opts.CustomerID != "" {
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		}
	}

	if opts.TruckerID != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.TruckerID,
			},
		}
	}

	data := map[string]any{
		"type":       "cost-codes",
		"attributes": attributes,
	}

	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Post(cmd.Context(), "/v1/cost-codes", jsonBody)
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

	row := costCodeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created cost code %s (%s)\n", row.ID, row.Code)
	return nil
}

func parseDoCostCodesCreateOptions(cmd *cobra.Command) (doCostCodesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	description, _ := cmd.Flags().GetString("description")
	isActive, _ := cmd.Flags().GetBool("active")
	customerID, _ := cmd.Flags().GetString("customer")
	truckerID, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostCodesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Code:        code,
		Description: description,
		IsActive:    isActive,
		CustomerID:  customerID,
		TruckerID:   truckerID,
	}, nil
}

func costCodeRowFromSingle(resp jsonAPISingleResponse) costCodeRow {
	return costCodeRow{
		ID:          resp.Data.ID,
		Code:        stringAttr(resp.Data.Attributes, "code"),
		Description: stringAttr(resp.Data.Attributes, "description"),
		IsActive:    boolAttr(resp.Data.Attributes, "is-active"),
	}
}
