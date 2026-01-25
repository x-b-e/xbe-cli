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

type integrationConfigsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type integrationConfigDetails struct {
	ID               string `json:"id"`
	FriendlyName     string `json:"friendly_name,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	BrokerName       string `json:"broker_name,omitempty"`
	Organization     string `json:"organization,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func newIntegrationConfigsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show integration config details",
		Long: `Show the full details of a specific integration config.

Output Fields:
  ID                 Integration config identifier
  NAME               Friendly name
  BROKER             Broker name or ID
  ORGANIZATION       Organization name or Type/ID
  CREATED AT         Creation timestamp
  UPDATED AT         Update timestamp

Arguments:
  <id>  Integration config ID (required). Find IDs using the list command.`,
		Example: `  # Show an integration config by ID
  xbe view integration-configs show 123

  # Output as JSON
  xbe view integration-configs show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIntegrationConfigsShow,
	}
	initIntegrationConfigsShowFlags(cmd)
	return cmd
}

func init() {
	integrationConfigsCmd.AddCommand(newIntegrationConfigsShowCmd())
}

func initIntegrationConfigsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIntegrationConfigsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIntegrationConfigsShowOptions(cmd)
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
		return fmt.Errorf("integration config id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[integration-configs]", strings.Join([]string{
		"friendly-name",
		"created-at",
		"updated-at",
		"broker",
		"organization",
	}, ","))
	query.Set("include", "broker,organization")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developers]", "name")
	query.Set("fields[material-suppliers]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/integration-configs/"+id, query)
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

	details := buildIntegrationConfigDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIntegrationConfigDetails(cmd, details)
}

func parseIntegrationConfigsShowOptions(cmd *cobra.Command) (integrationConfigsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return integrationConfigsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIntegrationConfigDetails(resp jsonAPISingleResponse) integrationConfigDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := integrationConfigDetails{
		ID:           resp.Data.ID,
		FriendlyName: stringAttr(attrs, "friendly-name"),
		CreatedAt:    formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:    formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Organization = organizationNameFromIncluded(org)
		}
	}

	return details
}

func renderIntegrationConfigDetails(cmd *cobra.Command, details integrationConfigDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FriendlyName != "" {
		fmt.Fprintf(out, "Name: %s\n", details.FriendlyName)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	orgLabel := formatRelated(details.Organization, formatPolymorphic(details.OrganizationType, details.OrganizationID))
	if orgLabel != "" {
		fmt.Fprintf(out, "Organization: %s\n", orgLabel)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
