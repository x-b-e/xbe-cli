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

type rawTransportTractorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawTransportTractorDetails struct {
	ID                string   `json:"id"`
	ExternalTractorID string   `json:"external_tractor_id,omitempty"`
	Importer          string   `json:"importer,omitempty"`
	ImportStatus      string   `json:"import_status,omitempty"`
	ImportErrors      []string `json:"import_errors,omitempty"`
	Tables            any      `json:"tables,omitempty"`
	BrokerID          string   `json:"broker_id,omitempty"`
	BrokerName        string   `json:"broker_name,omitempty"`
	TractorID         string   `json:"tractor_id,omitempty"`
	TractorNumber     string   `json:"tractor_number,omitempty"`
}

func newRawTransportTractorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw transport tractor details",
		Long: `Show the full details of a raw transport tractor.

Output Fields:
  ID                Raw transport tractor identifier
  External Tractor  External tractor identifier
  Importer          Importer name
  Import Status     Import status
  Import Errors     Import error messages
  Broker            Broker name or ID
  Tractor           Tractor number or ID
  Tables            Raw table payloads

Arguments:
  <id>  Raw transport tractor ID (required).`,
		Example: `  # Show raw transport tractor details
  xbe view raw-transport-tractors show 123

  # Output as JSON
  xbe view raw-transport-tractors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawTransportTractorsShow,
	}
	initRawTransportTractorsShowFlags(cmd)
	return cmd
}

func init() {
	rawTransportTractorsCmd.AddCommand(newRawTransportTractorsShowCmd())
}

func initRawTransportTractorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportTractorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawTransportTractorsShowOptions(cmd)
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
		return fmt.Errorf("raw transport tractor id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-transport-tractors]", "external-tractor-id,importer,import-status,import-errors,tables,broker,tractor")
	query.Set("include", "broker,tractor")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[tractors]", "number")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-tractors/"+id, query)
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

	details := buildRawTransportTractorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawTransportTractorDetails(cmd, details)
}

func parseRawTransportTractorsShowOptions(cmd *cobra.Command) (rawTransportTractorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportTractorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawTransportTractorDetails(resp jsonAPISingleResponse) rawTransportTractorDetails {
	row := rawTransportTractorRowFromSingle(resp)
	details := rawTransportTractorDetails{
		ID:                row.ID,
		ExternalTractorID: row.ExternalTractorID,
		Importer:          row.Importer,
		ImportStatus:      row.ImportStatus,
		BrokerID:          row.BrokerID,
		BrokerName:        row.BrokerName,
		TractorID:         row.TractorID,
		TractorNumber:     row.TractorNumber,
	}

	details.ImportErrors = stringSliceAttr(resp.Data.Attributes, "import-errors")
	details.Tables = anyAttr(resp.Data.Attributes, "tables")

	return details
}

func renderRawTransportTractorDetails(cmd *cobra.Command, details rawTransportTractorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalTractorID != "" {
		fmt.Fprintf(out, "External Tractor ID: %s\n", details.ExternalTractorID)
	}
	if details.Importer != "" {
		fmt.Fprintf(out, "Importer: %s\n", details.Importer)
	}
	if details.ImportStatus != "" {
		fmt.Fprintf(out, "Import Status: %s\n", details.ImportStatus)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.TractorID != "" || details.TractorNumber != "" {
		fmt.Fprintf(out, "Tractor: %s\n", formatRelated(details.TractorNumber, details.TractorID))
	}
	if len(details.ImportErrors) > 0 {
		fmt.Fprintf(out, "Import Errors: %s\n", strings.Join(details.ImportErrors, "; "))
	}

	if details.Tables != nil {
		fmt.Fprintf(out, "Tables: %d\n", countTables(details.Tables))
		if formatted := formatAnyJSON(details.Tables); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Table Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}

func countTables(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}
