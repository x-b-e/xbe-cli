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

type rawRecordsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawRecordDetails struct {
	ID                                string `json:"id"`
	ExternalRecordType                string `json:"external_record_type,omitempty"`
	ExternalRecordID                  string `json:"external_record_id,omitempty"`
	InternalRecordType                string `json:"internal_record_type,omitempty"`
	InternalRecordID                  string `json:"internal_record_id,omitempty"`
	BrokerID                          string `json:"broker_id,omitempty"`
	BrokerName                        string `json:"broker_name,omitempty"`
	IntegrationConfigID               string `json:"integration_config_id,omitempty"`
	IntegrationConfigName             string `json:"integration_config_name,omitempty"`
	IntegrationConfigOrganizationType string `json:"integration_config_organization_type,omitempty"`
	IntegrationConfigOrganizationID   string `json:"integration_config_organization_id,omitempty"`
	IsProcessed                       bool   `json:"is_processed,omitempty"`
	IsFailed                          bool   `json:"is_failed,omitempty"`
	IsSkipped                         bool   `json:"is_skipped,omitempty"`
	ProcessStartAt                    string `json:"process_start_at,omitempty"`
	ProcessEndAt                      string `json:"process_end_at,omitempty"`
	ProcessFailedAt                   string `json:"process_failed_at,omitempty"`
	ProcessDetail                     any    `json:"process_detail,omitempty"`
	InternalLinkages                  any    `json:"internal_linkages,omitempty"`
	RawAttributes                     any    `json:"raw_attributes,omitempty"`
}

func newRawRecordsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw record details",
		Long: `Show the full details of a specific ingest raw record.

Output Fields:
  ID                      Raw record identifier
  External Record         External record type and ID
  Internal Record         Internal record type and ID
  Broker                  Broker name or ID
  Integration Config      Integration config name or ID
  Integration Config Org  Integration config organization (Type/ID)
  Processed               Processed status
  Failed                  Failed status
  Skipped                 Skipped status
  Process Start At        Process start timestamp
  Process End At          Process end timestamp
  Process Failed At       Process failed timestamp
  Process Detail          Process detail payload
  Internal Linkages       Internal linkage data
  Raw Attributes          Raw payload attributes

Arguments:
  <id>  Raw record ID (required). Find IDs using the list command.`,
		Example: `  # Show raw record details
  xbe view raw-records show 123

  # Output as JSON
  xbe view raw-records show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawRecordsShow,
	}
	initRawRecordsShowFlags(cmd)
	return cmd
}

func init() {
	rawRecordsCmd.AddCommand(newRawRecordsShowCmd())
}

func initRawRecordsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawRecordsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawRecordsShowOptions(cmd)
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
		return fmt.Errorf("raw record id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-records]", "external-record-type,external-record-id,internal-record-type,internal-record-id,internal-linkages,raw-attributes,is-processed,is-failed,is-skipped,process-detail,process-start-at,process-end-at,process-failed-at,broker,integration-config")
	query.Set("include", "broker,integration-config")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[integration-configs]", "friendly-name,organization")

	body, _, err := client.Get(cmd.Context(), "/v1/ingest/raw-records/"+id, query)
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

	details := buildRawRecordDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawRecordDetails(cmd, details)
}

func parseRawRecordsShowOptions(cmd *cobra.Command) (rawRecordsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawRecordsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawRecordDetails(resp jsonAPISingleResponse) rawRecordDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := rawRecordDetails{
		ID:                 resp.Data.ID,
		ExternalRecordType: stringAttr(attrs, "external-record-type"),
		ExternalRecordID:   stringAttr(attrs, "external-record-id"),
		InternalRecordType: stringAttr(attrs, "internal-record-type"),
		InternalRecordID:   stringAttr(attrs, "internal-record-id"),
		IsProcessed:        boolAttr(attrs, "is-processed"),
		IsFailed:           boolAttr(attrs, "is-failed"),
		IsSkipped:          boolAttr(attrs, "is-skipped"),
		ProcessStartAt:     formatDateTime(stringAttr(attrs, "process-start-at")),
		ProcessEndAt:       formatDateTime(stringAttr(attrs, "process-end-at")),
		ProcessFailedAt:    formatDateTime(stringAttr(attrs, "process-failed-at")),
		ProcessDetail:      anyAttr(attrs, "process-detail"),
		InternalLinkages:   anyAttr(attrs, "internal-linkages"),
		RawAttributes:      anyAttr(attrs, "raw-attributes"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resp.Data.Relationships["integration-config"]; ok && rel.Data != nil {
		details.IntegrationConfigID = rel.Data.ID
		if config, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.IntegrationConfigName = stringAttr(config.Attributes, "friendly-name")
			if orgRel, ok := config.Relationships["organization"]; ok && orgRel.Data != nil {
				details.IntegrationConfigOrganizationType = orgRel.Data.Type
				details.IntegrationConfigOrganizationID = orgRel.Data.ID
			}
		}
	}

	return details
}

func renderRawRecordDetails(cmd *cobra.Command, details rawRecordDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalRecordType != "" || details.ExternalRecordID != "" {
		fmt.Fprintf(out, "External Record: %s\n", formatPolymorphic(details.ExternalRecordType, details.ExternalRecordID))
	}
	if details.InternalRecordType != "" || details.InternalRecordID != "" {
		fmt.Fprintf(out, "Internal Record: %s\n", formatPolymorphic(details.InternalRecordType, details.InternalRecordID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.IntegrationConfigID != "" || details.IntegrationConfigName != "" {
		fmt.Fprintf(out, "Integration Config: %s\n", formatRelated(details.IntegrationConfigName, details.IntegrationConfigID))
	}
	if details.IntegrationConfigOrganizationType != "" || details.IntegrationConfigOrganizationID != "" {
		fmt.Fprintf(out, "Integration Config Org: %s\n", formatPolymorphic(details.IntegrationConfigOrganizationType, details.IntegrationConfigOrganizationID))
	}

	fmt.Fprintf(out, "Processed: %t\n", details.IsProcessed)
	fmt.Fprintf(out, "Failed: %t\n", details.IsFailed)
	fmt.Fprintf(out, "Skipped: %t\n", details.IsSkipped)

	if details.ProcessStartAt != "" {
		fmt.Fprintf(out, "Process Start At: %s\n", details.ProcessStartAt)
	}
	if details.ProcessEndAt != "" {
		fmt.Fprintf(out, "Process End At: %s\n", details.ProcessEndAt)
	}
	if details.ProcessFailedAt != "" {
		fmt.Fprintf(out, "Process Failed At: %s\n", details.ProcessFailedAt)
	}

	if details.ProcessDetail != nil {
		switch typed := details.ProcessDetail.(type) {
		case string:
			typed = strings.TrimSpace(typed)
			if typed != "" {
				fmt.Fprintf(out, "Process Detail: %s\n", typed)
			}
		default:
			if formatted := formatAnyJSON(details.ProcessDetail); formatted != "" {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "Process Detail:")
				fmt.Fprintln(out, formatted)
			}
		}
	}

	if details.InternalLinkages != nil {
		if formatted := formatAnyJSON(details.InternalLinkages); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Internal Linkages:")
			fmt.Fprintln(out, formatted)
		}
	}

	if details.RawAttributes != nil {
		if formatted := formatAnyJSON(details.RawAttributes); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Raw Attributes:")
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
