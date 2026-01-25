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

type projectPhaseRevenueItemActualExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseRevenueItemActualExportDetails struct {
	ID                         string   `json:"id"`
	Status                     string   `json:"status,omitempty"`
	FileName                   string   `json:"file_name,omitempty"`
	RevenueDate                string   `json:"revenue_date,omitempty"`
	Body                       string   `json:"body,omitempty"`
	MimeType                   string   `json:"mime_type,omitempty"`
	FormatterErrorsDetails     any      `json:"formatter_errors_details,omitempty"`
	OrganizationType           string   `json:"organization_type,omitempty"`
	OrganizationID             string   `json:"organization_id,omitempty"`
	BrokerID                   string   `json:"broker_id,omitempty"`
	OrganizationFormatterID    string   `json:"organization_formatter_id,omitempty"`
	CreatedByID                string   `json:"created_by_id,omitempty"`
	ProjectPhaseRevenueItemIDs []string `json:"project_phase_revenue_item_ids,omitempty"`
}

func newProjectPhaseRevenueItemActualExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase revenue item actual export details",
		Long: `Show the full details of a project phase revenue item actual export.

Output Fields:
  ID
  Status
  File Name
  Revenue Date
  Body
  Mime Type
  Formatter Errors Details
  Organization (type + ID)
  Broker ID
  Organization Formatter ID
  Created By
  Project Phase Revenue Item IDs

Arguments:
  <id>    The export ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an export
  xbe view project-phase-revenue-item-actual-exports show 123

  # JSON output
  xbe view project-phase-revenue-item-actual-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseRevenueItemActualExportsShow,
	}
	initProjectPhaseRevenueItemActualExportsShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemActualExportsCmd.AddCommand(newProjectPhaseRevenueItemActualExportsShowCmd())
}

func initProjectPhaseRevenueItemActualExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemActualExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectPhaseRevenueItemActualExportsShowOptions(cmd)
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
		return fmt.Errorf("export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-revenue-item-actual-exports]", "body,file-name,formatter-errors-details,mime-type,status,revenue-date,broker,created-by,organization,organization-formatter,project-phase-revenue-items")
	query.Set("fields[project-phase-revenue-items]", "project-phase,project-revenue-item")
	query.Set("include", "project-phase-revenue-items")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-item-actual-exports/"+id, query)
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

	details := buildProjectPhaseRevenueItemActualExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseRevenueItemActualExportDetails(cmd, details)
}

func parseProjectPhaseRevenueItemActualExportsShowOptions(cmd *cobra.Command) (projectPhaseRevenueItemActualExportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return projectPhaseRevenueItemActualExportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return projectPhaseRevenueItemActualExportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return projectPhaseRevenueItemActualExportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return projectPhaseRevenueItemActualExportsShowOptions{}, err
	}

	return projectPhaseRevenueItemActualExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseRevenueItemActualExportDetails(resp jsonAPISingleResponse) projectPhaseRevenueItemActualExportDetails {
	resource := resp.Data
	attrs := resource.Attributes

	projectPhaseRevenueItemIDs := relationshipIDsFromMap(resource.Relationships, "project-phase-revenue-items")
	if len(projectPhaseRevenueItemIDs) == 0 && len(resp.Included) > 0 {
		for _, included := range resp.Included {
			if included.Type == "project-phase-revenue-items" && included.ID != "" {
				projectPhaseRevenueItemIDs = append(projectPhaseRevenueItemIDs, included.ID)
			}
		}
	}

	details := projectPhaseRevenueItemActualExportDetails{
		ID:                         resource.ID,
		Status:                     stringAttr(attrs, "status"),
		FileName:                   stringAttr(attrs, "file-name"),
		RevenueDate:                stringAttr(attrs, "revenue-date"),
		Body:                       stringAttr(attrs, "body"),
		MimeType:                   stringAttr(attrs, "mime-type"),
		FormatterErrorsDetails:     attrs["formatter-errors-details"],
		BrokerID:                   relationshipIDFromMap(resource.Relationships, "broker"),
		OrganizationFormatterID:    relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:                relationshipIDFromMap(resource.Relationships, "created-by"),
		ProjectPhaseRevenueItemIDs: projectPhaseRevenueItemIDs,
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	return details
}

func renderProjectPhaseRevenueItemActualExportDetails(cmd *cobra.Command, details projectPhaseRevenueItemActualExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
	}
	if details.RevenueDate != "" {
		fmt.Fprintf(out, "Revenue Date: %s\n", details.RevenueDate)
	}
	if details.MimeType != "" {
		fmt.Fprintf(out, "Mime Type: %s\n", details.MimeType)
	}
	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s:%s\n", details.OrganizationType, details.OrganizationID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.OrganizationFormatterID != "" {
		fmt.Fprintf(out, "Organization Formatter: %s\n", details.OrganizationFormatterID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}

	if details.Body != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Body:")
		fmt.Fprintln(out, details.Body)
	}
	if details.FormatterErrorsDetails != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Formatter Errors Details:")
		fmt.Fprintln(out, formatJSONBlock(details.FormatterErrorsDetails, "  "))
	}
	if len(details.ProjectPhaseRevenueItemIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Revenue Item IDs: %s\n", strings.Join(details.ProjectPhaseRevenueItemIDs, ", "))
	}

	return nil
}
