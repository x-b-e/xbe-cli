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

type ticketReportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type ticketReportDetails struct {
	ID                                 string   `json:"id"`
	FileName                           string   `json:"file_name,omitempty"`
	BrokerID                           string   `json:"broker_id,omitempty"`
	TicketReportTypeID                 string   `json:"ticket_report_type_id,omitempty"`
	TransformResult                    any      `json:"transform_result,omitempty"`
	TransformError                     string   `json:"transform_error,omitempty"`
	TimeCardIDs                        []string `json:"time_card_ids,omitempty"`
	MaterialTransactionTicketReportIDs []string `json:"material_transaction_ticket_report_ids,omitempty"`
	TicketReportDispatchIDs            []string `json:"ticket_report_dispatch_ids,omitempty"`
	TicketReportImportIDs              []string `json:"ticket_report_import_ids,omitempty"`
}

func newTicketReportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show ticket report details",
		Long: `Show the full details of a ticket report.

Output Fields:
  ID
  File Name
  Broker
  Ticket Report Type
  Transform Error
  Transform Result
  Time Card IDs
  Material Transaction Ticket Report IDs
  Ticket Report Dispatch IDs
  Ticket Report Import IDs

Arguments:
  <id>    The ticket report ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a ticket report
  xbe view ticket-reports show 123

  # JSON output
  xbe view ticket-reports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTicketReportsShow,
	}
	initTicketReportsShowFlags(cmd)
	return cmd
}

func init() {
	ticketReportsCmd.AddCommand(newTicketReportsShowCmd())
}

func initTicketReportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTicketReportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTicketReportsShowOptions(cmd)
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
		return fmt.Errorf("ticket report id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[ticket-reports]", "file-name,transform-result,transform-error,time-card-ids,material-transaction-ticket-report-ids,broker,ticket-report-type,ticket-report-dispatches,ticket-report-imports")

	body, _, err := client.Get(cmd.Context(), "/v1/ticket-reports/"+id, query)
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

	details := buildTicketReportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTicketReportDetails(cmd, details)
}

func parseTicketReportsShowOptions(cmd *cobra.Command) (ticketReportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return ticketReportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return ticketReportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return ticketReportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return ticketReportsShowOptions{}, err
	}

	return ticketReportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTicketReportDetails(resp jsonAPISingleResponse) ticketReportDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return ticketReportDetails{
		ID:                                 resource.ID,
		FileName:                           stringAttr(attrs, "file-name"),
		BrokerID:                           relationshipIDFromMap(resource.Relationships, "broker"),
		TicketReportTypeID:                 relationshipIDFromMap(resource.Relationships, "ticket-report-type"),
		TransformResult:                    attrs["transform-result"],
		TransformError:                     stringAttr(attrs, "transform-error"),
		TimeCardIDs:                        stringSliceAttr(attrs, "time-card-ids"),
		MaterialTransactionTicketReportIDs: stringSliceAttr(attrs, "material-transaction-ticket-report-ids"),
		TicketReportDispatchIDs:            relationshipIDsFromMap(resource.Relationships, "ticket-report-dispatches"),
		TicketReportImportIDs:              relationshipIDsFromMap(resource.Relationships, "ticket-report-imports"),
	}
}

func renderTicketReportDetails(cmd *cobra.Command, details ticketReportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.TicketReportTypeID != "" {
		fmt.Fprintf(out, "Ticket Report Type: %s\n", details.TicketReportTypeID)
	}
	if details.TransformError != "" {
		fmt.Fprintf(out, "Transform Error: %s\n", details.TransformError)
	}
	if len(details.TimeCardIDs) > 0 {
		fmt.Fprintf(out, "Time Card IDs: %s\n", strings.Join(details.TimeCardIDs, ", "))
	}
	if len(details.MaterialTransactionTicketReportIDs) > 0 {
		fmt.Fprintf(out, "Material Transaction Ticket Report IDs: %s\n", strings.Join(details.MaterialTransactionTicketReportIDs, ", "))
	}
	if len(details.TicketReportDispatchIDs) > 0 {
		fmt.Fprintf(out, "Ticket Report Dispatch IDs: %s\n", strings.Join(details.TicketReportDispatchIDs, ", "))
	}
	if len(details.TicketReportImportIDs) > 0 {
		fmt.Fprintf(out, "Ticket Report Import IDs: %s\n", strings.Join(details.TicketReportImportIDs, ", "))
	}
	if details.TransformResult != nil {
		fmt.Fprintln(out, "\nTransform Result:")
		fmt.Fprintln(out, formatJSONBlock(details.TransformResult, "  "))
	}

	return nil
}
