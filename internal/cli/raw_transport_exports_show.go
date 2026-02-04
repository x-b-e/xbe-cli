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

type rawTransportExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawTransportExportDetails struct {
	ID                   string   `json:"id"`
	ExternalOrderNumber  string   `json:"external_order_number,omitempty"`
	TargetDatabase       string   `json:"target_database,omitempty"`
	TargetTable          string   `json:"target_table,omitempty"`
	ExportType           string   `json:"export_type,omitempty"`
	Headers              any      `json:"headers,omitempty"`
	Rows                 any      `json:"rows,omitempty"`
	FormattedExport      string   `json:"formatted_export,omitempty"`
	Checksum             string   `json:"checksum,omitempty"`
	Sequence             string   `json:"sequence,omitempty"`
	StpNumbers           []string `json:"stp_numbers,omitempty"`
	IsExportable         bool     `json:"is_exportable,omitempty"`
	IsExported           bool     `json:"is_exported,omitempty"`
	NotExportableReasons []string `json:"not_exportable_reasons,omitempty"`
	IssueType            string   `json:"issue_type,omitempty"`
	FirstSeenAt          string   `json:"first_seen_at,omitempty"`
	ThrottledUntil       string   `json:"throttled_until,omitempty"`
	ExportResults        any      `json:"export_results,omitempty"`
	ExportedAt           string   `json:"exported_at,omitempty"`
	CreatedAt            string   `json:"created_at,omitempty"`
	BrokerID             string   `json:"broker_id,omitempty"`
	BrokerName           string   `json:"broker_name,omitempty"`
	TransportOrderID     string   `json:"transport_order_id,omitempty"`
	TransportOrderNumber string   `json:"transport_order_external_number,omitempty"`
	CreatedByID          string   `json:"created_by_id,omitempty"`
	CreatedByName        string   `json:"created_by_name,omitempty"`
}

func newRawTransportExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw transport export details",
		Long: `Show the full details of a raw transport export.

Output Fields:
  ID                    Raw transport export identifier
  External Order Number External order number
  Export Type           Export type
  Target Database       Target database
  Target Table          Target table
  Exportable            Exportable flag
  Exported              Exported flag
  Issue Type            Issue type
  Checksum              Export checksum
  Sequence              Export sequence
  STP Numbers           Stop numbers
  Not Exportable Reasons Non-exportable reasons
  First Seen At         First seen timestamp
  Throttled Until       Throttle expiration timestamp
  Exported At           Exported timestamp
  Created At            Created timestamp
  Broker                Broker name or ID
  Transport Order       Transport order external number or ID
  Created By            Creator user name or ID
  Headers               Header payload
  Rows                  Row payload
  Formatted Export      Formatted export payload
  Export Results        Export results payload

Arguments:
  <id>  Raw transport export ID (required).`,
		Example: `  # Show raw transport export details
  xbe view raw-transport-exports show 123

  # Output as JSON
  xbe view raw-transport-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawTransportExportsShow,
	}
	initRawTransportExportsShowFlags(cmd)
	return cmd
}

func init() {
	rawTransportExportsCmd.AddCommand(newRawTransportExportsShowCmd())
}

func initRawTransportExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportExportsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseRawTransportExportsShowOptions(cmd)
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
		return fmt.Errorf("raw transport export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-transport-exports]", strings.Join([]string{
		"external-order-number",
		"target-database",
		"target-table",
		"export-type",
		"headers",
		"rows",
		"formatted-export",
		"checksum",
		"sequence",
		"stp-numbers",
		"is-exportable",
		"is-exported",
		"not-exportable-reasons",
		"issue-type",
		"first-seen-at",
		"throttled-until",
		"export-results",
		"exported-at",
		"created-at",
		"broker",
		"transport-order",
		"created-by",
	}, ","))
	query.Set("include", "broker,transport-order,created-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[transport-orders]", "external-order-number")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-exports/"+id, query)
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

	details := buildRawTransportExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawTransportExportDetails(cmd, details)
}

func parseRawTransportExportsShowOptions(cmd *cobra.Command) (rawTransportExportsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawTransportExportDetails(resp jsonAPISingleResponse) rawTransportExportDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	row := buildRawTransportExportRow(resp.Data, included)
	attrs := resp.Data.Attributes
	details := rawTransportExportDetails{
		ID:                   row.ID,
		ExternalOrderNumber:  row.ExternalOrderNumber,
		TargetDatabase:       stringAttr(attrs, "target-database"),
		TargetTable:          stringAttr(attrs, "target-table"),
		ExportType:           row.ExportType,
		Headers:              anyAttr(attrs, "headers"),
		Rows:                 anyAttr(attrs, "rows"),
		FormattedExport:      stringAttr(attrs, "formatted-export"),
		Checksum:             stringAttr(attrs, "checksum"),
		Sequence:             stringAttr(attrs, "sequence"),
		StpNumbers:           stringSliceAttr(attrs, "stp-numbers"),
		IsExportable:         row.IsExportable,
		IsExported:           row.IsExported,
		NotExportableReasons: stringSliceAttr(attrs, "not-exportable-reasons"),
		IssueType:            row.IssueType,
		FirstSeenAt:          formatDateTime(stringAttr(attrs, "first-seen-at")),
		ThrottledUntil:       formatDateTime(stringAttr(attrs, "throttled-until")),
		ExportResults:        anyAttr(attrs, "export-results"),
		ExportedAt:           formatDateTime(stringAttr(attrs, "exported-at")),
		CreatedAt:            formatDateTime(stringAttr(attrs, "created-at")),
		BrokerID:             row.BrokerID,
		BrokerName:           row.BrokerName,
		TransportOrderID:     row.TransportOrderID,
		TransportOrderNumber: row.TransportOrderNumber,
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(user.Attributes, "name")
		}
	}

	return details
}

func renderRawTransportExportDetails(cmd *cobra.Command, details rawTransportExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalOrderNumber != "" {
		fmt.Fprintf(out, "External Order Number: %s\n", details.ExternalOrderNumber)
	}
	if details.ExportType != "" {
		fmt.Fprintf(out, "Export Type: %s\n", details.ExportType)
	}
	if details.TargetDatabase != "" {
		fmt.Fprintf(out, "Target Database: %s\n", details.TargetDatabase)
	}
	if details.TargetTable != "" {
		fmt.Fprintf(out, "Target Table: %s\n", details.TargetTable)
	}
	fmt.Fprintf(out, "Exportable: %t\n", details.IsExportable)
	fmt.Fprintf(out, "Exported: %t\n", details.IsExported)
	if details.IssueType != "" {
		fmt.Fprintf(out, "Issue Type: %s\n", details.IssueType)
	}
	if details.Checksum != "" {
		fmt.Fprintf(out, "Checksum: %s\n", details.Checksum)
	}
	if details.Sequence != "" {
		fmt.Fprintf(out, "Sequence: %s\n", details.Sequence)
	}
	if len(details.StpNumbers) > 0 {
		fmt.Fprintf(out, "STP Numbers: %s\n", strings.Join(details.StpNumbers, "; "))
	}
	if len(details.NotExportableReasons) > 0 {
		fmt.Fprintf(out, "Not Exportable Reasons: %s\n", strings.Join(details.NotExportableReasons, "; "))
	}
	if details.FirstSeenAt != "" {
		fmt.Fprintf(out, "First Seen At: %s\n", details.FirstSeenAt)
	}
	if details.ThrottledUntil != "" {
		fmt.Fprintf(out, "Throttled Until: %s\n", details.ThrottledUntil)
	}
	if details.ExportedAt != "" {
		fmt.Fprintf(out, "Exported At: %s\n", details.ExportedAt)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.TransportOrderID != "" || details.TransportOrderNumber != "" {
		fmt.Fprintf(out, "Transport Order: %s\n", formatRelated(details.TransportOrderNumber, details.TransportOrderID))
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}

	printRawTransportExportPayload(out, "Headers", details.Headers)
	printRawTransportExportPayload(out, "Rows", details.Rows)
	printRawTransportExportPayload(out, "Export Results", details.ExportResults)

	if details.FormattedExport != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Formatted Export:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.FormattedExport)
	}

	return nil
}

func printRawTransportExportPayload(out io.Writer, label string, payload any) {
	if payload == nil {
		return
	}

	formatted := formatAnyJSON(payload)
	if formatted == "" {
		return
	}

	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "%s:\n", label)
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintln(out, formatted)
}
