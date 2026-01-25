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

type doOpenDoorIssuesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Description  string
	Status       string
	Organization string
	ReportedBy   string
}

func newDoOpenDoorIssuesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an open door issue",
		Long: `Create a new open door issue.

Required flags:
  --description   Issue description (required)
  --status        Issue status (required: editing, reported, resolved)
  --organization  Organization in Type|ID format (required, e.g. Broker|123)
  --reported-by   Reporting user ID (required)

Optional flags:
  --json          Output JSON`,
		Example: `  # Create an open door issue
  xbe do open-door-issues create \
    --description "Driver reported a safety concern" \
    --status editing \
    --organization "Broker|123" \
    --reported-by 456

  # Create and return JSON
  xbe do open-door-issues create \
    --description "Concern with site access" \
    --status reported \
    --organization "Customer|789" \
    --reported-by 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoOpenDoorIssuesCreate,
	}
	initDoOpenDoorIssuesCreateFlags(cmd)
	return cmd
}

func init() {
	doOpenDoorIssuesCmd.AddCommand(newDoOpenDoorIssuesCreateCmd())
}

func initDoOpenDoorIssuesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Issue description (required)")
	cmd.Flags().String("status", "", "Issue status (required: editing, reported, resolved)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (required, e.g. Broker|123)")
	cmd.Flags().String("reported-by", "", "Reporting user ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenDoorIssuesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOpenDoorIssuesCreateOptions(cmd)
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

	if opts.Description == "" {
		err := fmt.Errorf("--description is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Status == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Organization == "" {
		err := fmt.Errorf("--organization is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ReportedBy == "" {
		err := fmt.Errorf("--reported-by is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	orgType, orgID, err := parseOrganization(opts.Organization)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"description": opts.Description,
		"status":      opts.Status,
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		},
		"reported-by": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ReportedBy,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "open-door-issues",
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

	body, _, err := client.Post(cmd.Context(), "/v1/open-door-issues", jsonBody)
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

	row := buildOpenDoorIssueRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created open door issue %s\n", row.ID)
	return nil
}

func parseDoOpenDoorIssuesCreateOptions(cmd *cobra.Command) (doOpenDoorIssuesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	organization, _ := cmd.Flags().GetString("organization")
	reportedBy, _ := cmd.Flags().GetString("reported-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenDoorIssuesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Description:  description,
		Status:       status,
		Organization: organization,
		ReportedBy:   reportedBy,
	}, nil
}

func buildOpenDoorIssueRowFromSingle(resp jsonAPISingleResponse) openDoorIssueRow {
	resource := resp.Data
	row := openDoorIssueRow{
		ID:          resource.ID,
		Status:      strings.TrimSpace(stringAttr(resource.Attributes, "status")),
		Description: strings.TrimSpace(stringAttr(resource.Attributes, "description")),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationID = rel.Data.ID
		row.OrganizationType = rel.Data.Type
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Organization = strings.TrimSpace(stringAttr(org.Attributes, "company-name"))
			if row.Organization == "" {
				row.Organization = strings.TrimSpace(stringAttr(org.Attributes, "name"))
			}
		}
	}

	if rel, ok := resource.Relationships["reported-by"]; ok && rel.Data != nil {
		row.ReportedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ReportedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return row
}
