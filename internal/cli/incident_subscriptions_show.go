package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type incidentSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentSubscriptionDetails struct {
	ID                      string `json:"id"`
	UserID                  string `json:"user_id,omitempty"`
	Kind                    string `json:"kind,omitempty"`
	ContactMethod           string `json:"contact_method,omitempty"`
	CalculatedContactMethod string `json:"calculated_contact_method,omitempty"`
	OrganizationType        string `json:"organization_type,omitempty"`
	OrganizationID          string `json:"organization_id,omitempty"`
	CustomerID              string `json:"customer_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	MaterialSupplierID      string `json:"material_supplier_id,omitempty"`
	IncidentType            string `json:"incident_type,omitempty"`
	IncidentID              string `json:"incident_id,omitempty"`
}

func newIncidentSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident subscription details",
		Long: `Show full details of an incident subscription.

Output Fields:
  ID                        Subscription identifier
  User                      User ID
  Kind                      Incident kind filter
  Contact Method            Explicit contact method
  Calculated Contact Method Effective contact method
  Organization              Organization scope (type/id)
  Customer                  Customer ID (if scoped to a customer)
  Broker                    Broker ID (if scoped to a broker)
  Material Supplier         Material supplier ID (if scoped to a material supplier)
  Incident                  Incident scope (type/id)

Arguments:
  <id>    Incident subscription ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an incident subscription
  xbe view incident-subscriptions show 123

  # JSON output
  xbe view incident-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentSubscriptionsShow,
	}
	initIncidentSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	incidentSubscriptionsCmd.AddCommand(newIncidentSubscriptionsShowCmd())
}

func initIncidentSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIncidentSubscriptionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("incident subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-subscriptions]", "kind,contact-method,calculated-contact-method,user,organization,customer,broker,material-supplier,incident")

	body, _, err := client.Get(cmd.Context(), "/v1/incident-subscriptions/"+id, query)
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

	details := buildIncidentSubscriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentSubscriptionDetails(cmd, details)
}

func parseIncidentSubscriptionsShowOptions(cmd *cobra.Command) (incidentSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentSubscriptionDetails(resp jsonAPISingleResponse) incidentSubscriptionDetails {
	row := buildIncidentSubscriptionRow(resp.Data)
	return incidentSubscriptionDetails{
		ID:                      row.ID,
		UserID:                  row.UserID,
		Kind:                    row.Kind,
		ContactMethod:           row.ContactMethod,
		CalculatedContactMethod: row.CalculatedContactMethod,
		OrganizationType:        row.OrganizationType,
		OrganizationID:          row.OrganizationID,
		CustomerID:              row.CustomerID,
		BrokerID:                row.BrokerID,
		MaterialSupplierID:      row.MaterialSupplierID,
		IncidentType:            row.IncidentType,
		IncidentID:              row.IncidentID,
	}
}

func renderIncidentSubscriptionDetails(cmd *cobra.Command, details incidentSubscriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "User: %s\n", formatOptional(details.UserID))
	fmt.Fprintf(out, "Kind: %s\n", formatOptional(details.Kind))
	fmt.Fprintf(out, "Contact Method: %s\n", formatOptional(details.ContactMethod))
	fmt.Fprintf(out, "Calculated Contact Method: %s\n", formatOptional(details.CalculatedContactMethod))
	fmt.Fprintf(out, "Organization: %s\n", formatOptional(incidentSubscriptionRelationshipDisplay(details.OrganizationType, details.OrganizationID)))
	fmt.Fprintf(out, "Customer: %s\n", formatOptional(details.CustomerID))
	fmt.Fprintf(out, "Broker: %s\n", formatOptional(details.BrokerID))
	fmt.Fprintf(out, "Material Supplier: %s\n", formatOptional(details.MaterialSupplierID))
	fmt.Fprintf(out, "Incident: %s\n", formatOptional(incidentSubscriptionRelationshipDisplay(details.IncidentType, details.IncidentID)))

	return nil
}
