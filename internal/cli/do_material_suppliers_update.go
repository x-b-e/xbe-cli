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

type doMaterialSuppliersUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	Name                 string
	URL                  string
	PhoneNumber          string
	IsActive             bool
	IsControlledByBroker bool
}

func newDoMaterialSuppliersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material supplier",
		Long: `Update a material supplier.

Optional flags:
  --name                    Material supplier name
  --url                     Company website URL
  --phone-number            Contact phone number
  --active                  Set as active
  --no-active               Set as inactive
  --is-controlled-by-broker Supplier is controlled by broker
  --no-is-controlled-by-broker Supplier is not controlled by broker`,
		Example: `  # Update material supplier name
  xbe do material-suppliers update 123 --name "New Name"

  # Deactivate a material supplier
  xbe do material-suppliers update 123 --no-active

  # Update URL
  xbe do material-suppliers update 123 --url "https://newsite.com"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSuppliersUpdate,
	}
	initDoMaterialSuppliersUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSuppliersCmd.AddCommand(newDoMaterialSuppliersUpdateCmd())
}

func initDoMaterialSuppliersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Material supplier name")
	cmd.Flags().String("url", "", "Company website URL")
	cmd.Flags().String("phone-number", "", "Contact phone number")
	cmd.Flags().Bool("active", false, "Set as active")
	cmd.Flags().Bool("no-active", false, "Set as inactive")
	cmd.Flags().Bool("is-controlled-by-broker", false, "Supplier is controlled by broker")
	cmd.Flags().Bool("no-is-controlled-by-broker", false, "Supplier is not controlled by broker")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSuppliersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSuppliersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("url") {
		attributes["url"] = opts.URL
	}
	if cmd.Flags().Changed("phone-number") {
		attributes["phone-number"] = opts.PhoneNumber
	}
	if cmd.Flags().Changed("active") {
		attributes["is-active"] = true
	}
	if cmd.Flags().Changed("no-active") {
		attributes["is-active"] = false
	}
	if cmd.Flags().Changed("is-controlled-by-broker") {
		attributes["is-controlled-by-broker"] = true
	}
	if cmd.Flags().Changed("no-is-controlled-by-broker") {
		attributes["is-controlled-by-broker"] = false
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "material-suppliers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/material-suppliers/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material supplier %s\n", row.ID)
	return nil
}

func parseDoMaterialSuppliersUpdateOptions(cmd *cobra.Command, args []string) (doMaterialSuppliersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	isActive, _ := cmd.Flags().GetBool("active")
	isControlledByBroker, _ := cmd.Flags().GetBool("is-controlled-by-broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSuppliersUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		Name:                 name,
		URL:                  url,
		PhoneNumber:          phoneNumber,
		IsActive:             isActive,
		IsControlledByBroker: isControlledByBroker,
	}, nil
}
