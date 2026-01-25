package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type invoiceRevisionizingWorksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type invoiceRevisionizingWorkDetails struct {
	ID                   string `json:"id"`
	Comment              string `json:"comment,omitempty"`
	FormattedResultsText string `json:"formatted_results_text,omitempty"`
	Results              any    `json:"results,omitempty"`
	Messages             any    `json:"messages,omitempty"`
	IsRetry              bool   `json:"is_retry"`
	JID                  string `json:"jid,omitempty"`
	ScheduledAt          string `json:"scheduled_at,omitempty"`
	ProcessedAt          string `json:"processed_at,omitempty"`
	WorkResults          any    `json:"work_results,omitempty"`
	WorkErrors           any    `json:"work_errors,omitempty"`

	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	BrokerName       string `json:"broker_name,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	CreatedByName    string `json:"created_by_name,omitempty"`
}

func newInvoiceRevisionizingWorksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show invoice revisionizing work details",
		Long: `Show full details of an invoice revisionizing work item.

Includes metadata, results, messages, and related references.

Arguments:
  <id>  The revisionizing work ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show invoice revisionizing work details
  xbe view invoice-revisionizing-works show 123

  # Output as JSON
  xbe view invoice-revisionizing-works show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoiceRevisionizingWorksShow,
	}
	initInvoiceRevisionizingWorksShowFlags(cmd)
	return cmd
}

func init() {
	invoiceRevisionizingWorksCmd.AddCommand(newInvoiceRevisionizingWorksShowCmd())
}

func initInvoiceRevisionizingWorksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceRevisionizingWorksShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInvoiceRevisionizingWorksShowOptions(cmd)
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
		return fmt.Errorf("invoice revisionizing work id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[invoice-revisionizing-works]", "comment,formatted-results-text,results,messages,is-retry,jid,scheduled-at,processed-at,work-results,work-errors,organization,broker,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("include", "organization,broker,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-revisionizing-works/"+id, query)
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

	details := buildInvoiceRevisionizingWorkDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInvoiceRevisionizingWorkDetails(cmd, details)
}

func parseInvoiceRevisionizingWorksShowOptions(cmd *cobra.Command) (invoiceRevisionizingWorksShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceRevisionizingWorksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInvoiceRevisionizingWorkDetails(resp jsonAPISingleResponse) invoiceRevisionizingWorkDetails {
	attrs := resp.Data.Attributes
	details := invoiceRevisionizingWorkDetails{
		ID:                   resp.Data.ID,
		Comment:              strings.TrimSpace(stringAttr(attrs, "comment")),
		FormattedResultsText: strings.TrimSpace(stringAttr(attrs, "formatted-results-text")),
		Results:              anyAttr(attrs, "results"),
		Messages:             anyAttr(attrs, "messages"),
		IsRetry:              boolAttr(attrs, "is-retry"),
		JID:                  stringAttr(attrs, "jid"),
		ScheduledAt:          formatDateTime(stringAttr(attrs, "scheduled-at")),
		ProcessedAt:          formatDateTime(stringAttr(attrs, "processed-at")),
		WorkResults:          anyAttr(attrs, "work-results"),
		WorkErrors:           anyAttr(attrs, "work-errors"),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OrganizationName = firstNonEmpty(
				stringAttr(inc.Attributes, "company-name"),
				stringAttr(inc.Attributes, "name"),
			)
		}
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(inc.Attributes, "company-name"))
		}
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	return details
}

func renderInvoiceRevisionizingWorkDetails(cmd *cobra.Command, details invoiceRevisionizingWorkDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	if details.JID != "" {
		fmt.Fprintf(out, "JID: %s\n", details.JID)
	}
	if details.ScheduledAt != "" {
		fmt.Fprintf(out, "Scheduled At: %s\n", details.ScheduledAt)
	}
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	fmt.Fprintf(out, "Is Retry: %s\n", formatBool(details.IsRetry))

	if orgLabel := formatPolymorphicLabel(details.OrganizationType, details.OrganizationName, details.OrganizationID); orgLabel != "" {
		fmt.Fprintf(out, "Organization: %s\n", orgLabel)
	}
	writeLabelWithID(out, "Broker", details.BrokerName, details.BrokerID)
	writeLabelWithID(out, "Created By", details.CreatedByName, details.CreatedByID)

	if details.FormattedResultsText != "" {
		fmt.Fprintln(out, "Formatted Results Text:")
		fmt.Fprintln(out, indentLines(details.FormattedResultsText, "  "))
	}

	writeAnySection(out, "Results", details.Results)
	writeAnySection(out, "Messages", details.Messages)
	writeAnySection(out, "Work Results", details.WorkResults)
	writeAnySection(out, "Work Errors", details.WorkErrors)

	return nil
}

func writeAnySection(out io.Writer, label string, value any) {
	fmt.Fprintf(out, "%s:\n", label)
	formatted := formatAny(value)
	if formatted == "" {
		fmt.Fprintln(out, "  (none)")
		return
	}
	fmt.Fprintln(out, indentLines(formatted, "  "))
}

func formatPolymorphicLabel(kind, name, id string) string {
	kind = strings.TrimSpace(kind)
	name = strings.TrimSpace(name)
	id = strings.TrimSpace(id)

	switch {
	case kind != "" && name != "" && id != "":
		return fmt.Sprintf("%s %s (%s)", kind, name, id)
	case kind != "" && id != "":
		return fmt.Sprintf("%s %s", kind, id)
	case kind != "" && name != "":
		return fmt.Sprintf("%s %s", kind, name)
	case name != "" && id != "":
		return fmt.Sprintf("%s (%s)", name, id)
	case name != "":
		return name
	case kind != "":
		return kind
	case id != "":
		return id
	default:
		return ""
	}
}
