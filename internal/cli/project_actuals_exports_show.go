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

type projectActualsExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectActualsExportDetails struct {
	ID                      string   `json:"id"`
	Status                  string   `json:"status,omitempty"`
	FileName                string   `json:"file_name,omitempty"`
	Body                    string   `json:"body,omitempty"`
	MimeType                string   `json:"mime_type,omitempty"`
	FormatterErrorsDetails  any      `json:"formatter_errors_details,omitempty"`
	OrganizationType        string   `json:"organization_type,omitempty"`
	OrganizationID          string   `json:"organization_id,omitempty"`
	BrokerID                string   `json:"broker_id,omitempty"`
	ProjectID               string   `json:"project_id,omitempty"`
	OrganizationFormatterID string   `json:"organization_formatter_id,omitempty"`
	CreatedByID             string   `json:"created_by_id,omitempty"`
	JobProductionPlanIDs    []string `json:"job_production_plan_ids,omitempty"`
}

func newProjectActualsExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project actuals export details",
		Long: `Show the full details of a project actuals export.

Output Fields:
  ID
  Status
  File Name
  Body
  Mime Type
  Formatter Errors Details
  Organization (type + ID)
  Broker ID
  Project ID
  Organization Formatter ID
  Created By
  Job Production Plan IDs

Arguments:
  <id>    The export ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an export
  xbe view project-actuals-exports show 123

  # JSON output
  xbe view project-actuals-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectActualsExportsShow,
	}
	initProjectActualsExportsShowFlags(cmd)
	return cmd
}

func init() {
	projectActualsExportsCmd.AddCommand(newProjectActualsExportsShowCmd())
}

func initProjectActualsExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectActualsExportsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectActualsExportsShowOptions(cmd)
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
	query.Set("fields[project-actuals-exports]", "body,file-name,formatter-errors-details,mime-type,status,broker,project,created-by,organization,organization-formatter,job-production-plans")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("include", "job-production-plans")

	body, _, err := client.Get(cmd.Context(), "/v1/project-actuals-exports/"+id, query)
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

	details := buildProjectActualsExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectActualsExportDetails(cmd, details)
}

func parseProjectActualsExportsShowOptions(cmd *cobra.Command) (projectActualsExportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return projectActualsExportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return projectActualsExportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return projectActualsExportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return projectActualsExportsShowOptions{}, err
	}

	return projectActualsExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectActualsExportDetails(resp jsonAPISingleResponse) projectActualsExportDetails {
	resource := resp.Data
	attrs := resource.Attributes

	jobProductionPlanIDs := relationshipIDsFromMap(resource.Relationships, "job-production-plans")
	if len(jobProductionPlanIDs) == 0 && len(resp.Included) > 0 {
		for _, included := range resp.Included {
			if included.Type == "job-production-plans" && included.ID != "" {
				jobProductionPlanIDs = append(jobProductionPlanIDs, included.ID)
			}
		}
	}

	details := projectActualsExportDetails{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		FileName:                stringAttr(attrs, "file-name"),
		Body:                    stringAttr(attrs, "body"),
		MimeType:                stringAttr(attrs, "mime-type"),
		FormatterErrorsDetails:  attrs["formatter-errors-details"],
		BrokerID:                relationshipIDFromMap(resource.Relationships, "broker"),
		ProjectID:               relationshipIDFromMap(resource.Relationships, "project"),
		OrganizationFormatterID: relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:             relationshipIDFromMap(resource.Relationships, "created-by"),
		JobProductionPlanIDs:    jobProductionPlanIDs,
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	return details
}

func renderProjectActualsExportDetails(cmd *cobra.Command, details projectActualsExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
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
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project: %s\n", details.ProjectID)
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
	if len(details.JobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan IDs: %s\n", strings.Join(details.JobProductionPlanIDs, ", "))
	}

	return nil
}
