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

type projectTransportOrganizationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportOrganizationDetails struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name"`
	ExternalTmsMasterCompanyID  string   `json:"external_tms_master_company_id,omitempty"`
	BrokerID                    string   `json:"broker_id,omitempty"`
	BrokerName                  string   `json:"broker_name,omitempty"`
	ProjectTransportLocationIDs []string `json:"project_transport_location_ids,omitempty"`
	ExternalIdentificationIDs   []string `json:"external_identification_ids,omitempty"`
}

func newProjectTransportOrganizationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport organization details",
		Long: `Show the full details of a project transport organization.

Output Fields:
  ID
  Name
  External TMS Master Company ID
  Broker
  Project Transport Locations
  External Identifications

Arguments:
  <id>    The project transport organization ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport organization
  xbe view project-transport-organizations show 123

  # Output as JSON
  xbe view project-transport-organizations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportOrganizationsShow,
	}
	initProjectTransportOrganizationsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportOrganizationsCmd.AddCommand(newProjectTransportOrganizationsShowCmd())
}

func initProjectTransportOrganizationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportOrganizationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportOrganizationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport organization id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-organizations]", "name,external-tms-master-company-id,broker,project-transport-locations,external-identifications")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-organizations/"+id, query)
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

	details := buildProjectTransportOrganizationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportOrganizationDetails(cmd, details)
}

func parseProjectTransportOrganizationsShowOptions(cmd *cobra.Command) (projectTransportOrganizationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportOrganizationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportOrganizationDetails(resp jsonAPISingleResponse) projectTransportOrganizationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := projectTransportOrganizationDetails{
		ID:                         resource.ID,
		Name:                       stringAttr(attrs, "name"),
		ExternalTmsMasterCompanyID: stringAttr(attrs, "external-tms-master-company-id"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}
	if rel, ok := resource.Relationships["project-transport-locations"]; ok {
		details.ProjectTransportLocationIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDList(rel)
	}

	return details
}

func renderProjectTransportOrganizationDetails(cmd *cobra.Command, details projectTransportOrganizationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.ExternalTmsMasterCompanyID != "" {
		fmt.Fprintf(out, "External TMS Master Company ID: %s\n", details.ExternalTmsMasterCompanyID)
	}
	if details.BrokerID != "" {
		if details.BrokerName != "" {
			fmt.Fprintf(out, "Broker: %s (%s)\n", details.BrokerName, details.BrokerID)
		} else {
			fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
		}
	}

	if len(details.ProjectTransportLocationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Project Transport Locations (%d):\n", len(details.ProjectTransportLocationIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.ProjectTransportLocationIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "External Identifications (%d):\n", len(details.ExternalIdentificationIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.ExternalIdentificationIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
