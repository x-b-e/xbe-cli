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

type doChangeRequestsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Requests         string
	Organization     string
	OrganizationType string
	OrganizationID   string
}

func newDoChangeRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a change request",
		Long: `Create a change request.

Optional flags:
  --requests           Request items as JSON array (use [] for empty)
  --organization       Organization in Type|ID format (e.g., Broker|123)
  --organization-type  Organization type (optional if --organization is set)
  --organization-id    Organization ID (optional if --organization is set)`,
		Example: `  # Create a change request for a broker
  xbe do change-requests create \
    --requests '[{"field":"status","from":"draft","to":"approved"}]' \
    --organization-type brokers \
    --organization-id 123

  # Create a change request with Type|ID organization format
  xbe do change-requests create \
    --requests '[{"field":"note","value":"Update"}]' \
    --organization "Broker|123"

  # Output as JSON
  xbe do change-requests create \
    --requests '[{"field":"status","from":"draft","to":"approved"}]' \
    --organization "Broker|123" \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoChangeRequestsCreate,
	}
	initDoChangeRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doChangeRequestsCmd.AddCommand(newDoChangeRequestsCreateCmd())
}

func initDoChangeRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("requests", "", "Request items as JSON array (use [] for empty)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (e.g., Broker|123)")
	cmd.Flags().String("organization-type", "", "Organization type (optional if --organization is set)")
	cmd.Flags().String("organization-id", "", "Organization ID (optional if --organization is set)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoChangeRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoChangeRequestsCreateOptions(cmd)
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

	orgType, orgID, err := resolveChangeRequestOrganization(cmd, opts.Organization, opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("requests") {
		requests, err := parseChangeRequestRequests(opts.Requests)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["requests"] = requests
	}

	relationships := map[string]any{}
	if orgType != "" && orgID != "" {
		relationships["organization"] = map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		}
	}

	requestData := map[string]any{
		"type":       "change-requests",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/change-requests", jsonBody)
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

	row := buildChangeRequestRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created change request %s\n", row.ID)
	return nil
}

func parseDoChangeRequestsCreateOptions(cmd *cobra.Command) (doChangeRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	requests, _ := cmd.Flags().GetString("requests")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doChangeRequestsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Requests:         requests,
		Organization:     organization,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}

func buildChangeRequestRowFromSingle(resp jsonAPISingleResponse) changeRequestRow {
	row := changeRequestRow{
		ID:            resp.Data.ID,
		RequestsCount: requestCountFromAny(resp.Data.Attributes["requests"]),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func resolveChangeRequestOrganization(cmd *cobra.Command, organization, orgType, orgID string) (string, string, error) {
	if cmd.Flags().Changed("organization") {
		return parseOrganization(organization)
	}
	if cmd.Flags().Changed("organization-type") || cmd.Flags().Changed("organization-id") {
		if strings.TrimSpace(orgType) == "" || strings.TrimSpace(orgID) == "" {
			return "", "", fmt.Errorf("--organization-type and --organization-id must be provided together")
		}
		return parseOrganization(fmt.Sprintf("%s|%s", orgType, orgID))
	}
	return "", "", nil
}

func parseChangeRequestRequests(raw string) ([]map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("--requests must be a JSON array (use [] for empty)")
	}
	var data []map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("invalid --requests JSON: %w", err)
	}
	return data, nil
}
