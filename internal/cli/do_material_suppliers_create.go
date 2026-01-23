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

type doMaterialSuppliersCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	Name                 string
	BrokerID             string
	URL                  string
	PhoneNumber          string
	IsActive             bool
	IsControlledByBroker bool
}

func newDoMaterialSuppliersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new material supplier",
		Long: `Create a new material supplier.

Required flags:
  --name      Material supplier name
  --broker    Broker ID (required)

Optional flags:
  --url                     Company website URL
  --phone-number            Contact phone number
  --active                  Set as active (default: true)
  --is-controlled-by-broker Supplier is controlled by broker`,
		Example: `  # Create a basic material supplier
  xbe do material-suppliers create --name "Acme Materials" --broker 123

  # Create with contact info
  xbe do material-suppliers create --name "Quality Aggregates" --broker 123 \
    --url "https://qualityagg.com" --phone-number "555-1234"`,
		RunE: runDoMaterialSuppliersCreate,
	}
	initDoMaterialSuppliersCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSuppliersCmd.AddCommand(newDoMaterialSuppliersCreateCmd())
}

func initDoMaterialSuppliersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Material supplier name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("url", "", "Company website URL")
	cmd.Flags().String("phone-number", "", "Contact phone number")
	cmd.Flags().Bool("active", true, "Set as active")
	cmd.Flags().Bool("is-controlled-by-broker", false, "Supplier is controlled by broker")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("broker")
}

func runDoMaterialSuppliersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSuppliersCreateOptions(cmd)
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
		"name":      opts.Name,
		"is-active": opts.IsActive,
	}

	if opts.URL != "" {
		attributes["url"] = opts.URL
	}
	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if cmd.Flags().Changed("is-controlled-by-broker") {
		attributes["is-controlled-by-broker"] = opts.IsControlledByBroker
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-suppliers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-suppliers", jsonBody)
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

	row := materialSupplierRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material supplier %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoMaterialSuppliersCreateOptions(cmd *cobra.Command) (doMaterialSuppliersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	brokerID, _ := cmd.Flags().GetString("broker")
	url, _ := cmd.Flags().GetString("url")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	isActive, _ := cmd.Flags().GetBool("active")
	isControlledByBroker, _ := cmd.Flags().GetBool("is-controlled-by-broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSuppliersCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		Name:                 name,
		BrokerID:             brokerID,
		URL:                  url,
		PhoneNumber:          phoneNumber,
		IsActive:             isActive,
		IsControlledByBroker: isControlledByBroker,
	}, nil
}

func materialSupplierRowFromSingle(resp jsonAPISingleResponse) materialSupplierRow {
	return materialSupplierRow{
		ID:       resp.Data.ID,
		Name:     stringAttr(resp.Data.Attributes, "name"),
		IsActive: boolAttr(resp.Data.Attributes, "is-active"),
	}
}
