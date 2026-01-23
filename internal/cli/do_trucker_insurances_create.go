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

type doTruckerInsurancesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Trucker     string
	CompanyName string
	ContactName string
	PhoneNumber string
}

func newDoTruckerInsurancesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trucker insurance",
		Long: `Create a trucker insurance.

Required:
  --trucker       Trucker ID

Optional:
  --company-name  Insurance company name
  --contact-name  Contact name
  --phone-number  Phone number`,
		Example: `  # Create a trucker insurance
  xbe do trucker-insurances create --trucker 123 --company-name "ABC Insurance"

  # Create with all details
  xbe do trucker-insurances create --trucker 123 --company-name "ABC Insurance" \
    --contact-name "John Doe" --phone-number "555-123-4567"`,
		RunE: runDoTruckerInsurancesCreate,
	}
	initDoTruckerInsurancesCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerInsurancesCmd.AddCommand(newDoTruckerInsurancesCreateCmd())
}

func initDoTruckerInsurancesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("company-name", "", "Insurance company name")
	cmd.Flags().String("contact-name", "", "Contact name")
	cmd.Flags().String("phone-number", "", "Phone number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("trucker")
}

func runDoTruckerInsurancesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerInsurancesCreateOptions(cmd)
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

	if opts.CompanyName != "" {
		attributes["company-name"] = opts.CompanyName
	}
	if opts.ContactName != "" {
		attributes["contact-name"] = opts.ContactName
	}
	if opts.PhoneNumber != "" {
		attributes["phone-number"] = opts.PhoneNumber
	}

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trucker-insurances",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-insurances", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker insurance %s\n", resp.Data.ID)
	return nil
}

func parseDoTruckerInsurancesCreateOptions(cmd *cobra.Command) (doTruckerInsurancesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	companyName, _ := cmd.Flags().GetString("company-name")
	contactName, _ := cmd.Flags().GetString("contact-name")
	phoneNumber, _ := cmd.Flags().GetString("phone-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerInsurancesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Trucker:     trucker,
		CompanyName: companyName,
		ContactName: contactName,
		PhoneNumber: phoneNumber,
	}, nil
}
