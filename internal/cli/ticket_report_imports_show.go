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

type ticketReportImportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type ticketReportImportDetails struct {
	ID                    string `json:"id"`
	TicketReportID        string `json:"ticket_report_id,omitempty"`
	TicketReportFileName  string `json:"ticket_report_file_name,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	BrokerName            string `json:"broker_name,omitempty"`
	Status                string `json:"status,omitempty"`
	ImportResults         any    `json:"import_results,omitempty"`
	ImportErrors          any    `json:"import_errors,omitempty"`
	ImportWarnings        any    `json:"import_warnings,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
	CurrentUserCanDestroy bool   `json:"current_user_can_destroy"`
	CanDelete             bool   `json:"can_delete"`
}

func newTicketReportImportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show ticket report import details",
		Long: `Show the full details of a ticket report import.

Output Fields:
  ID                         Ticket report import identifier
  Ticket Report              Ticket report file name and ID
  Broker                     Broker name and ID
  Status                     Import status
  Created At                 Created timestamp
  Updated At                 Updated timestamp
  Import Results             Imported record results
  Import Errors              Import errors
  Import Warnings            Import warnings
  Current User Can Destroy   Permission to delete
  Can Delete                 Delete eligibility

Arguments:
  <id>  The ticket report import ID (required).`,
		Example: `  # Show a ticket report import
  xbe view ticket-report-imports show 123

  # Output as JSON
  xbe view ticket-report-imports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTicketReportImportsShow,
	}
	initTicketReportImportsShowFlags(cmd)
	return cmd
}

func init() {
	ticketReportImportsCmd.AddCommand(newTicketReportImportsShowCmd())
}

func initTicketReportImportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTicketReportImportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTicketReportImportsShowOptions(cmd)
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
		return fmt.Errorf("ticket report import id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[ticket-report-imports]", "status,import-results,import-errors,import-warnings,created-at,updated-at,broker,ticket-report")
	query.Set("include", "broker,ticket-report")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[ticket-reports]", "file-name")
	query.Set("meta[ticket-report-import]", "current_user_can_destroy,can_delete")

	body, _, err := client.Get(cmd.Context(), "/v1/ticket-report-imports/"+id, query)
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

	details := buildTicketReportImportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTicketReportImportDetails(cmd, details)
}

func parseTicketReportImportsShowOptions(cmd *cobra.Command) (ticketReportImportsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return ticketReportImportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTicketReportImportDetails(resp jsonAPISingleResponse) ticketReportImportDetails {
	row := ticketReportImportRowFromSingle(resp)
	attrs := resp.Data.Attributes

	details := ticketReportImportDetails{
		ID:                    row.ID,
		TicketReportID:        row.TicketReportID,
		TicketReportFileName:  row.TicketReportFileName,
		BrokerID:              row.BrokerID,
		BrokerName:            row.BrokerName,
		Status:                stringAttr(attrs, "status"),
		ImportResults:         anyAttr(attrs, "import-results"),
		ImportErrors:          anyAttr(attrs, "import-errors"),
		ImportWarnings:        anyAttr(attrs, "import-warnings"),
		CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
		CurrentUserCanDestroy: boolAttr(resp.Data.Meta, "current_user_can_destroy"),
		CanDelete:             boolAttr(resp.Data.Meta, "can_delete"),
	}

	return details
}

func renderTicketReportImportDetails(cmd *cobra.Command, details ticketReportImportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TicketReportID != "" || details.TicketReportFileName != "" {
		fmt.Fprintf(out, "Ticket Report: %s\n", formatRelated(details.TicketReportFileName, details.TicketReportID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	fmt.Fprintf(out, "Current User Can Destroy: %t\n", details.CurrentUserCanDestroy)
	fmt.Fprintf(out, "Can Delete: %t\n", details.CanDelete)

	if formatted := formatAnyJSON(details.ImportResults); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Import Results:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	if formatted := formatAnyJSON(details.ImportErrors); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Import Errors:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	if formatted := formatAnyJSON(details.ImportWarnings); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Import Warnings:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	return nil
}
