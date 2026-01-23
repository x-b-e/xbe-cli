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

type doTruckerInsurancesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	CompanyName string
	ContactName string
	PhoneNumber string
}

func newDoTruckerInsurancesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker insurance",
		Long: `Update a trucker insurance.

Optional:
  --company-name  Insurance company name
  --contact-name  Contact name
  --phone-number  Phone number`,
		Example: `  # Update company name
  xbe do trucker-insurances update 123 --company-name "XYZ Insurance"

  # Update contact info
  xbe do trucker-insurances update 123 --contact-name "Jane Smith" --phone-number "555-987-6543"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerInsurancesUpdate,
	}
	initDoTruckerInsurancesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerInsurancesCmd.AddCommand(newDoTruckerInsurancesUpdateCmd())
}

func initDoTruckerInsurancesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("company-name", "", "Insurance company name")
	cmd.Flags().String("contact-name", "", "Contact name")
	cmd.Flags().String("phone-number", "", "Phone number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerInsurancesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerInsurancesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("company-name") {
		attributes["company-name"] = opts.CompanyName
	}
	if cmd.Flags().Changed("contact-name") {
		attributes["contact-name"] = opts.ContactName
	}
	if cmd.Flags().Changed("phone-number") {
		attributes["phone-number"] = opts.PhoneNumber
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "trucker-insurances",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-insurances/"+opts.ID, jsonBody)
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
		row := truckerInsuranceRow{
			ID:          resp.Data.ID,
			CompanyName: stringAttr(resp.Data.Attributes, "company-name"),
			ContactName: stringAttr(resp.Data.Attributes, "contact-name"),
			PhoneNumber: stringAttr(resp.Data.Attributes, "phone-number"),
		}
		if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker insurance %s\n", resp.Data.ID)
	return nil
}

func parseDoTruckerInsurancesUpdateOptions(cmd *cobra.Command, args []string) (doTruckerInsurancesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	companyName, _ := cmd.Flags().GetString("company-name")
	contactName, _ := cmd.Flags().GetString("contact-name")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerInsurancesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		CompanyName: companyName,
		ContactName: contactName,
		PhoneNumber: phoneNumber,
	}, nil
}
