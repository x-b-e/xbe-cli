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

type timeSheetsExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetsExportDetails struct {
	ID                           string   `json:"id"`
	Status                       string   `json:"status,omitempty"`
	FileName                     string   `json:"file_name,omitempty"`
	MimeType                     string   `json:"mime_type,omitempty"`
	Body                         string   `json:"body,omitempty"`
	FormatterErrorsDetails       any      `json:"formatter_errors_details,omitempty"`
	OrganizationFormatterID      string   `json:"organization_formatter_id,omitempty"`
	OrganizationFormatterSummary string   `json:"organization_formatter_summary,omitempty"`
	Organization                 string   `json:"organization,omitempty"`
	OrganizationType             string   `json:"organization_type,omitempty"`
	OrganizationID               string   `json:"organization_id,omitempty"`
	BrokerID                     string   `json:"broker_id,omitempty"`
	BrokerName                   string   `json:"broker_name,omitempty"`
	CreatedByID                  string   `json:"created_by_id,omitempty"`
	CreatedByName                string   `json:"created_by_name,omitempty"`
	TimeSheetIDs                 []string `json:"time_sheet_ids,omitempty"`
	CreatedAt                    string   `json:"created_at,omitempty"`
	UpdatedAt                    string   `json:"updated_at,omitempty"`
}

func newTimeSheetsExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheets export details",
		Long: `Show the full details of a time sheets export.

Output Fields:
  ID                  Time sheets export identifier
  Status              Export status
  File Name            Export file name
  Mime Type            Export file MIME type
  Organization Formatter Organization formatter used for export
  Organization         Organization name or Type/ID
  Broker               Broker name or ID
  Created By           Creator user name or ID
  Time Sheets          Associated time sheet IDs
  Created At           Created timestamp
  Updated At           Updated timestamp
  Formatter Errors     Formatter error details
  Body                 Export body

Arguments:
  <id>  Time sheets export ID (required).`,
		Example: `  # Show time sheets export details
  xbe view time-sheets-exports show 123

  # Output as JSON
  xbe view time-sheets-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetsExportsShow,
	}
	initTimeSheetsExportsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetsExportsCmd.AddCommand(newTimeSheetsExportsShowCmd())
}

func initTimeSheetsExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetsExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetsExportsShowOptions(cmd)
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
		return fmt.Errorf("time sheets export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheets-exports]", strings.Join([]string{
		"status",
		"file-name",
		"mime-type",
		"body",
		"formatter-errors-details",
		"organization-formatter",
		"organization",
		"broker",
		"created-by",
		"time-sheets",
		"created-at",
		"updated-at",
	}, ","))
	query.Set("include", "organization-formatter,organization,broker,created-by")
	query.Set("fields[organization-formatters]", "description")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developers]", "name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheets-exports/"+id, query)
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

	details := buildTimeSheetsExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetsExportDetails(cmd, details)
}

func parseTimeSheetsExportsShowOptions(cmd *cobra.Command) (timeSheetsExportsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetsExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetsExportDetails(resp jsonAPISingleResponse) timeSheetsExportDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := timeSheetsExportDetails{
		ID:                     resp.Data.ID,
		Status:                 stringAttr(attrs, "status"),
		FileName:               stringAttr(attrs, "file-name"),
		MimeType:               stringAttr(attrs, "mime-type"),
		Body:                   stringAttr(attrs, "body"),
		FormatterErrorsDetails: anyAttr(attrs, "formatter-errors-details"),
		CreatedAt:              formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:              formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["organization-formatter"]; ok && rel.Data != nil {
		details.OrganizationFormatterID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OrganizationFormatterSummary = strings.TrimSpace(stringAttr(inc.Attributes, "description"))
		}
	}
	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Organization = organizationNameFromIncluded(inc)
		}
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = organizationNameFromIncluded(inc)
		}
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}
	if rel, ok := resp.Data.Relationships["time-sheets"]; ok {
		details.TimeSheetIDs = relationshipIDList(rel)
	}

	return details
}

func renderTimeSheetsExportDetails(cmd *cobra.Command, details timeSheetsExportDetails) error {
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
	if details.OrganizationFormatterID != "" || details.OrganizationFormatterSummary != "" {
		fmt.Fprintf(out, "Organization Formatter: %s\n", formatRelated(details.OrganizationFormatterSummary, details.OrganizationFormatterID))
	}
	orgLabel := formatRelated(details.Organization, formatPolymorphic(details.OrganizationType, details.OrganizationID))
	if orgLabel != "" {
		fmt.Fprintf(out, "Organization: %s\n", orgLabel)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}
	if len(details.TimeSheetIDs) > 0 {
		fmt.Fprintf(out, "Time Sheets: %s\n", strings.Join(details.TimeSheetIDs, ", "))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if formatted := formatAnyJSON(details.FormatterErrorsDetails); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Formatter Errors:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	if details.Body != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Body:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Body)
	}

	return nil
}
