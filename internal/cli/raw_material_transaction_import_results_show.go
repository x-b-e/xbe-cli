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

type rawMaterialTransactionImportResultsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawMaterialTransactionImportResultDetails struct {
	ID                           string   `json:"id"`
	Importer                     string   `json:"importer,omitempty"`
	ConfigurationID              string   `json:"configuration_id,omitempty"`
	LocationID                   string   `json:"location_id,omitempty"`
	TimeZoneID                   string   `json:"time_zone_id,omitempty"`
	ImportStart                  string   `json:"import_start,omitempty"`
	ImportEnd                    string   `json:"import_end,omitempty"`
	EarliestCreatedTransactionAt string   `json:"earliest_created_transaction_at,omitempty"`
	LatestCreatedTransactionAt   string   `json:"latest_created_transaction_at,omitempty"`
	HasErrors                    bool     `json:"has_errors"`
	BatchID                      string   `json:"batch_id,omitempty"`
	BeganAt                      string   `json:"began_at,omitempty"`
	EndedAt                      string   `json:"ended_at,omitempty"`
	DisconnectedAt               string   `json:"disconnected_at,omitempty"`
	LastConnectedAt              string   `json:"last_connected_at,omitempty"`
	IsConnected                  bool     `json:"is_connected"`
	Duplicates                   []string `json:"duplicates,omitempty"`
	Created                      []string `json:"created,omitempty"`
	Updated                      []string `json:"updated,omitempty"`
	ErrorMessages                []string `json:"error_messages,omitempty"`
	SourceType                   string   `json:"source_type,omitempty"`
	SourceID                     string   `json:"source_id,omitempty"`
	SourceName                   string   `json:"source,omitempty"`
	BrokerID                     string   `json:"broker_id,omitempty"`
	BrokerName                   string   `json:"broker,omitempty"`
}

func newRawMaterialTransactionImportResultsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw material transaction import result details",
		Long: `Show the full details of a raw material transaction import result.

Output Fields:
  ID
  Importer
  Configuration ID
  Location ID
  Time Zone
  Import Start
  Import End
  Earliest Created Transaction At
  Latest Created Transaction At
  Has Errors
  Batch ID
  Began At
  Ended At
  Disconnected At
  Last Connected At
  Is Connected
  Duplicates
  Created
  Updated
  Error Messages
  Source
  Broker

Arguments:
  <id>    The import result ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show import result details
  xbe view raw-material-transaction-import-results show 123

  # Output as JSON
  xbe view raw-material-transaction-import-results show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawMaterialTransactionImportResultsShow,
	}
	initRawMaterialTransactionImportResultsShowFlags(cmd)
	return cmd
}

func init() {
	rawMaterialTransactionImportResultsCmd.AddCommand(newRawMaterialTransactionImportResultsShowCmd())
}

func initRawMaterialTransactionImportResultsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawMaterialTransactionImportResultsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawMaterialTransactionImportResultsShowOptions(cmd)
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
		return fmt.Errorf("raw material transaction import result id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-material-transaction-import-results]", "importer,configurationid,locationid,time-zone-id,import-start,import-end,earliest-created-transaction-at,latest-created-transaction-at,has-errors,batch-id,began-at,ended-at,disconnected-at,last-connected-at,is-connected,duplicates,created,updated,error-messages")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "source,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-material-transaction-import-results/"+id, query)
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

	details := buildRawMaterialTransactionImportResultDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawMaterialTransactionImportResultDetails(cmd, details)
}

func parseRawMaterialTransactionImportResultsShowOptions(cmd *cobra.Command) (rawMaterialTransactionImportResultsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawMaterialTransactionImportResultsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawMaterialTransactionImportResultDetails(resp jsonAPISingleResponse) rawMaterialTransactionImportResultDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := rawMaterialTransactionImportResultDetails{
		ID:                           resource.ID,
		Importer:                     stringAttr(attrs, "importer"),
		ConfigurationID:              stringAttr(attrs, "configurationid"),
		LocationID:                   stringAttr(attrs, "locationid"),
		TimeZoneID:                   stringAttr(attrs, "time-zone-id"),
		ImportStart:                  formatDateTime(stringAttr(attrs, "import-start")),
		ImportEnd:                    formatDateTime(stringAttr(attrs, "import-end")),
		EarliestCreatedTransactionAt: formatDateTime(stringAttr(attrs, "earliest-created-transaction-at")),
		LatestCreatedTransactionAt:   formatDateTime(stringAttr(attrs, "latest-created-transaction-at")),
		HasErrors:                    boolAttr(attrs, "has-errors"),
		BatchID:                      stringAttr(attrs, "batch-id"),
		BeganAt:                      formatDateTime(stringAttr(attrs, "began-at")),
		EndedAt:                      formatDateTime(stringAttr(attrs, "ended-at")),
		DisconnectedAt:               formatDateTime(stringAttr(attrs, "disconnected-at")),
		LastConnectedAt:              formatDateTime(stringAttr(attrs, "last-connected-at")),
		IsConnected:                  boolAttr(attrs, "is-connected"),
		Duplicates:                   stringSliceAttr(attrs, "duplicates"),
		Created:                      stringSliceAttr(attrs, "created"),
		Updated:                      stringSliceAttr(attrs, "updated"),
		ErrorMessages:                stringSliceAttr(attrs, "error-messages"),
	}

	if rel, ok := resource.Relationships["source"]; ok && rel.Data != nil {
		details.SourceType = rel.Data.Type
		details.SourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.SourceID != "" && details.SourceType != "" {
		if src, ok := included[resourceKey(details.SourceType, details.SourceID)]; ok {
			details.SourceName = stringAttr(src.Attributes, "name")
		}
	}

	if details.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", details.BrokerID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	return details
}

func renderRawMaterialTransactionImportResultDetails(cmd *cobra.Command, details rawMaterialTransactionImportResultDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Importer != "" {
		fmt.Fprintf(out, "Importer: %s\n", details.Importer)
	}
	if details.ConfigurationID != "" {
		fmt.Fprintf(out, "Configuration ID: %s\n", details.ConfigurationID)
	}
	if details.LocationID != "" {
		fmt.Fprintf(out, "Location ID: %s\n", details.LocationID)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", details.TimeZoneID)
	}
	if details.ImportStart != "" {
		fmt.Fprintf(out, "Import Start: %s\n", details.ImportStart)
	}
	if details.ImportEnd != "" {
		fmt.Fprintf(out, "Import End: %s\n", details.ImportEnd)
	}
	if details.BeganAt != "" {
		fmt.Fprintf(out, "Began At: %s\n", details.BeganAt)
	}
	if details.EndedAt != "" {
		fmt.Fprintf(out, "Ended At: %s\n", details.EndedAt)
	}
	if details.EarliestCreatedTransactionAt != "" {
		fmt.Fprintf(out, "Earliest Created Transaction At: %s\n", details.EarliestCreatedTransactionAt)
	}
	if details.LatestCreatedTransactionAt != "" {
		fmt.Fprintf(out, "Latest Created Transaction At: %s\n", details.LatestCreatedTransactionAt)
	}
	if details.BatchID != "" {
		fmt.Fprintf(out, "Batch ID: %s\n", details.BatchID)
	}
	fmt.Fprintf(out, "Has Errors: %t\n", details.HasErrors)
	fmt.Fprintf(out, "Is Connected: %t\n", details.IsConnected)
	if details.LastConnectedAt != "" {
		fmt.Fprintf(out, "Last Connected At: %s\n", details.LastConnectedAt)
	}
	if details.DisconnectedAt != "" {
		fmt.Fprintf(out, "Disconnected At: %s\n", details.DisconnectedAt)
	}

	if details.SourceID != "" {
		label := details.SourceType + "/" + details.SourceID
		if details.SourceName != "" {
			label = fmt.Sprintf("%s (%s)", details.SourceName, label)
		}
		fmt.Fprintf(out, "Source: %s\n", label)
	}

	if details.BrokerID != "" {
		label := details.BrokerID
		if details.BrokerName != "" {
			label = fmt.Sprintf("%s (%s)", details.BrokerName, details.BrokerID)
		}
		fmt.Fprintf(out, "Broker: %s\n", label)
	}

	if len(details.Duplicates) > 0 {
		fmt.Fprintf(out, "Duplicates: %s\n", strings.Join(details.Duplicates, ", "))
	}
	if len(details.Created) > 0 {
		fmt.Fprintf(out, "Created: %s\n", strings.Join(details.Created, ", "))
	}
	if len(details.Updated) > 0 {
		fmt.Fprintf(out, "Updated: %s\n", strings.Join(details.Updated, ", "))
	}
	if len(details.ErrorMessages) > 0 {
		fmt.Fprintf(out, "Error Messages: %s\n", strings.Join(details.ErrorMessages, ", "))
	}

	return nil
}
