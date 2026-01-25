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

type changeRequestsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type changeRequestDetails struct {
	ID               string `json:"id"`
	Requests         any    `json:"requests,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	BrokerName       string `json:"broker_name,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	CreatedByName    string `json:"created_by_name,omitempty"`
}

func newChangeRequestsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show change request details",
		Long: `Show the full details of a specific change request.

Output Fields:
  ID            Change request identifier
  Organization  Organization type and ID
  Broker        Broker (name or ID)
  Created By    User who created the change request (name or ID)
  Requests      Request items (JSON)

Arguments:
  <id>          The change request ID (required).`,
		Example: `  # View a change request by ID
  xbe view change-requests show 123

  # Get change request as JSON
  xbe view change-requests show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runChangeRequestsShow,
	}
	initChangeRequestsShowFlags(cmd)
	return cmd
}

func init() {
	changeRequestsCmd.AddCommand(newChangeRequestsShowCmd())
}

func initChangeRequestsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runChangeRequestsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseChangeRequestsShowOptions(cmd)
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
		return fmt.Errorf("change request id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[change-requests]", "requests,organization,broker,created-by")
	query.Set("include", "broker,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/change-requests/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildChangeRequestDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderChangeRequestDetails(cmd, details)
}

func parseChangeRequestsShowOptions(cmd *cobra.Command) (changeRequestsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return changeRequestsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return changeRequestsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return changeRequestsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return changeRequestsShowOptions{}, err
	}

	return changeRequestsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildChangeRequestDetails(resp jsonAPISingleResponse) changeRequestDetails {
	details := changeRequestDetails{
		ID:       resp.Data.ID,
		Requests: anyAttr(resp.Data.Attributes, "requests"),
	}

	var brokerType string
	var createdByType string
	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		brokerType = rel.Data.Type
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.BrokerID != "" && brokerType != "" {
		if broker, ok := included[resourceKey(brokerType, details.BrokerID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if details.CreatedByID != "" && createdByType != "" {
		if user, ok := included[resourceKey(createdByType, details.CreatedByID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return details
}

func renderChangeRequestDetails(cmd *cobra.Command, details changeRequestDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s\n", formatPolymorphic(details.OrganizationType, details.OrganizationID))
	}

	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}

	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}

	if formatted := formatAnyJSON(details.Requests); formatted != "" {
		fmt.Fprintln(out, "Requests:")
		fmt.Fprintln(out, formatted)
	}

	return nil
}
