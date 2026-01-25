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

type invoiceGenerationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type invoiceGenerationDetails struct {
	ID                        string   `json:"id"`
	OrganizationType          string   `json:"organization_type,omitempty"`
	OrganizationID            string   `json:"organization_id,omitempty"`
	Organization              string   `json:"organization,omitempty"`
	CreatedByID               string   `json:"created_by_id,omitempty"`
	CreatedBy                 string   `json:"created_by,omitempty"`
	ParentInvoiceGenerationID string   `json:"parent_invoice_generation_id,omitempty"`
	ChildInvoiceGenerationIDs []string `json:"child_invoice_generation_ids,omitempty"`
	TimeCardIDs               []string `json:"time_card_ids,omitempty"`
	Note                      string   `json:"note,omitempty"`
	InvoicingDate             string   `json:"invoicing_date,omitempty"`
	CompletedAt               string   `json:"completed_at,omitempty"`
	IsRunning                 bool     `json:"is_running,omitempty"`
	IsParent                  bool     `json:"is_parent,omitempty"`
	IsCompleted               bool     `json:"is_completed,omitempty"`
	Status                    string   `json:"status,omitempty"`
	GenerationResults         any      `json:"generation_results,omitempty"`
	GenerationErrors          any      `json:"generation_errors,omitempty"`
	CreatedAt                 string   `json:"created_at,omitempty"`
	UpdatedAt                 string   `json:"updated_at,omitempty"`
}

func newInvoiceGenerationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show invoice generation details",
		Long: `Show the full details of an invoice generation.

Arguments:
  <id>  The invoice generation ID (required).`,
		Example: `  # Show an invoice generation
  xbe view invoice-generations show 123

  # Output as JSON
  xbe view invoice-generations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoiceGenerationsShow,
	}
	initInvoiceGenerationsShowFlags(cmd)
	return cmd
}

func init() {
	invoiceGenerationsCmd.AddCommand(newInvoiceGenerationsShowCmd())
}

func initInvoiceGenerationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceGenerationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInvoiceGenerationsShowOptions(cmd)
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
		return fmt.Errorf("invoice generation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[invoice-generations]", strings.Join([]string{
		"time-card-ids",
		"note",
		"invoicing-date",
		"completed-at",
		"is-running",
		"is-parent",
		"is-completed",
		"status",
		"generation-results",
		"generation-errors",
		"organization",
		"created-by",
		"parent-invoice-generation",
		"children",
		"created-at",
		"updated-at",
	}, ","))
	query.Set("include", "organization,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-generations/"+id, query)
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

	details := buildInvoiceGenerationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInvoiceGenerationDetails(cmd, details)
}

func parseInvoiceGenerationsShowOptions(cmd *cobra.Command) (invoiceGenerationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceGenerationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInvoiceGenerationDetails(resp jsonAPISingleResponse) invoiceGenerationDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := invoiceGenerationDetails{
		ID:                resp.Data.ID,
		TimeCardIDs:       stringSliceAttr(attrs, "time-card-ids"),
		Note:              stringAttr(attrs, "note"),
		InvoicingDate:     formatDate(stringAttr(attrs, "invoicing-date")),
		CompletedAt:       formatDateTime(stringAttr(attrs, "completed-at")),
		IsRunning:         boolAttr(attrs, "is-running"),
		IsParent:          boolAttr(attrs, "is-parent"),
		IsCompleted:       boolAttr(attrs, "is-completed"),
		Status:            stringAttr(attrs, "status"),
		GenerationResults: anyAttr(attrs, "generation-results"),
		GenerationErrors:  anyAttr(attrs, "generation-errors"),
		CreatedAt:         formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:         formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Organization = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["parent-invoice-generation"]; ok && rel.Data != nil {
		details.ParentInvoiceGenerationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["children"]; ok {
		details.ChildInvoiceGenerationIDs = relationshipIDList(rel)
	}

	return details
}

func renderInvoiceGenerationDetails(cmd *cobra.Command, details invoiceGenerationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OrganizationType != "" || details.OrganizationID != "" || details.Organization != "" {
		orgLabel := formatRelated(details.Organization, formatPolymorphic(details.OrganizationType, details.OrganizationID))
		fmt.Fprintf(out, "Organization: %s\n", orgLabel)
	}
	if details.CreatedByID != "" || details.CreatedBy != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedBy, details.CreatedByID))
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.InvoicingDate != "" {
		fmt.Fprintf(out, "Invoicing Date: %s\n", details.InvoicingDate)
	}
	if details.CompletedAt != "" {
		fmt.Fprintf(out, "Completed At: %s\n", details.CompletedAt)
	}
	fmt.Fprintf(out, "Is Completed: %t\n", details.IsCompleted)
	fmt.Fprintf(out, "Is Running: %t\n", details.IsRunning)
	fmt.Fprintf(out, "Is Parent: %t\n", details.IsParent)
	if len(details.TimeCardIDs) > 0 {
		fmt.Fprintf(out, "Time Cards: %s\n", strings.Join(details.TimeCardIDs, ", "))
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.ParentInvoiceGenerationID != "" {
		fmt.Fprintf(out, "Parent Generation: %s\n", details.ParentInvoiceGenerationID)
	}
	if len(details.ChildInvoiceGenerationIDs) > 0 {
		fmt.Fprintf(out, "Child Generations: %s\n", strings.Join(details.ChildInvoiceGenerationIDs, ", "))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if details.GenerationResults != nil {
		if formatted := formatAnyJSON(details.GenerationResults); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Generation Results:")
			fmt.Fprintln(out, formatted)
		}
	}

	if details.GenerationErrors != nil {
		if formatted := formatAnyJSON(details.GenerationErrors); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Generation Errors:")
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
