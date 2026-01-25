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

type organizationFormattersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationFormatterDetails struct {
	ID                string   `json:"id"`
	Description       string   `json:"description,omitempty"`
	FormatterType     string   `json:"formatter_type,omitempty"`
	Status            string   `json:"status,omitempty"`
	IsLibrary         bool     `json:"is_library"`
	HasFormatFunction bool     `json:"has_format_function"`
	MimeTypes         []string `json:"mime_types,omitempty"`
	FormatterFunction string   `json:"formatter_function,omitempty"`
	Organization      string   `json:"organization,omitempty"`
	OrganizationType  string   `json:"organization_type,omitempty"`
	OrganizationID    string   `json:"organization_id,omitempty"`
}

func newOrganizationFormattersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization formatter details",
		Long: `Show the full details of a specific organization formatter.

Output Fields:
  ID                 Formatter identifier
  Description        Formatter description
  Formatter Type     Formatter STI class name
  Status             Formatter status
  Library            Whether the formatter is a shared library
  Has Format Function Whether a format() function is defined
  Mime Types         Supported MIME types
  Organization       Organization name or Type/ID
  Formatter Function JavaScript formatter function source

Arguments:
  <id>    The organization formatter ID (required).`,
		Example: `  # Show formatter details
  xbe view organization-formatters show 123

  # Get JSON output
  xbe view organization-formatters show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationFormattersShow,
	}
	initOrganizationFormattersShowFlags(cmd)
	return cmd
}

func init() {
	organizationFormattersCmd.AddCommand(newOrganizationFormattersShowCmd())
}

func initOrganizationFormattersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationFormattersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOrganizationFormattersShowOptions(cmd)
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
		return fmt.Errorf("organization formatter id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-formatters]", "description,formatter-type,status,is-library,has-format-function,mime-types,formatter-function,organization")
	query.Set("include", "organization")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developers]", "name")
	query.Set("fields[material-suppliers]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-formatters/"+id, query)
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

	details := buildOrganizationFormatterDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationFormatterDetails(cmd, details)
}

func parseOrganizationFormattersShowOptions(cmd *cobra.Command) (organizationFormattersShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return organizationFormattersShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return organizationFormattersShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return organizationFormattersShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return organizationFormattersShowOptions{}, err
	}

	return organizationFormattersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationFormatterDetails(resp jsonAPISingleResponse) organizationFormatterDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := organizationFormatterDetails{
		ID:                resp.Data.ID,
		Description:       stringAttr(attrs, "description"),
		FormatterType:     stringAttr(attrs, "formatter-type"),
		Status:            stringAttr(attrs, "status"),
		IsLibrary:         boolAttr(attrs, "is-library"),
		HasFormatFunction: boolAttr(attrs, "has-format-function"),
		MimeTypes:         stringSliceAttr(attrs, "mime-types"),
		FormatterFunction: stringAttr(attrs, "formatter-function"),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationID = rel.Data.ID
		details.OrganizationType = rel.Data.Type
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Organization = organizationNameFromIncluded(inc)
		}
	}

	return details
}

func renderOrganizationFormatterDetails(cmd *cobra.Command, details organizationFormatterDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.FormatterType != "" {
		fmt.Fprintf(out, "Formatter Type: %s\n", details.FormatterType)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	fmt.Fprintf(out, "Library: %t\n", details.IsLibrary)
	fmt.Fprintf(out, "Has Format Function: %t\n", details.HasFormatFunction)
	if len(details.MimeTypes) > 0 {
		fmt.Fprintf(out, "Mime Types: %s\n", strings.Join(details.MimeTypes, ", "))
	}
	if details.OrganizationID != "" || details.OrganizationType != "" {
		orgLabel := formatRelated(details.Organization, formatPolymorphic(details.OrganizationType, details.OrganizationID))
		if orgLabel != "" {
			fmt.Fprintf(out, "Organization: %s\n", orgLabel)
		}
	}
	if details.FormatterFunction != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Formatter Function:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.FormatterFunction)
	}

	return nil
}
