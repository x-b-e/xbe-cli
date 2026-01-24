package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doRawTransportTrailersCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	BrokerID          string
	ExternalTrailerID string
	Importer          string
	Tables            string
}

func newDoRawTransportTrailersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw transport trailer",
		Long: `Create a raw transport trailer record.

Required flags:
  --broker               Broker ID
  --external-trailer-id  External trailer identifier
  --tables               Tables JSON array (raw transport payload)

Optional flags:
  --importer             Importer key (e.g., quantix_tmw)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a raw transport trailer with empty tables
  xbe do raw-transport-trailers create \
    --broker 123 \
    --external-trailer-id TRL-0001 \
    --importer quantix_tmw \
    --tables '[]'

  # Output as JSON
  xbe do raw-transport-trailers create --broker 123 --external-trailer-id TRL-0002 --tables '[]' --json`,
		Args: cobra.NoArgs,
		RunE: runDoRawTransportTrailersCreate,
	}
	initDoRawTransportTrailersCreateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportTrailersCmd.AddCommand(newDoRawTransportTrailersCreateCmd())
}

func initDoRawTransportTrailersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("external-trailer-id", "", "External trailer identifier (required)")
	cmd.Flags().String("tables", "", "Tables JSON array (required)")
	cmd.Flags().String("importer", "", "Importer key (e.g., quantix_tmw)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("external-trailer-id")
	cmd.MarkFlagRequired("tables")
}

func runDoRawTransportTrailersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawTransportTrailersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	externalTrailerID := strings.TrimSpace(opts.ExternalTrailerID)
	if externalTrailerID == "" {
		err := fmt.Errorf("--external-trailer-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	brokerID := strings.TrimSpace(opts.BrokerID)
	if brokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	tablesJSON := strings.TrimSpace(opts.Tables)
	if tablesJSON == "" {
		err := fmt.Errorf("--tables is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var parsed any
	if err := json.Unmarshal([]byte(tablesJSON), &parsed); err != nil {
		err := fmt.Errorf("invalid --tables JSON: %w", err)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	tables, ok := parsed.([]any)
	if !ok {
		err := fmt.Errorf("--tables must be a JSON array")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"external-trailer-id": externalTrailerID,
		"tables":              tables,
	}
	if strings.TrimSpace(opts.Importer) != "" {
		attributes["importer"] = strings.TrimSpace(opts.Importer)
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   brokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "raw-transport-trailers",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/raw-transport-trailers", jsonBody)
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

	row := buildRawTransportTrailerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport trailer %s\n", row.ID)
	return nil
}

func parseDoRawTransportTrailersCreateOptions(cmd *cobra.Command) (doRawTransportTrailersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	externalTrailerID, _ := cmd.Flags().GetString("external-trailer-id")
	importer, _ := cmd.Flags().GetString("importer")
	tables, _ := cmd.Flags().GetString("tables")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportTrailersCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		BrokerID:          brokerID,
		ExternalTrailerID: externalTrailerID,
		Importer:          importer,
		Tables:            tables,
	}, nil
}

func buildRawTransportTrailerRowFromSingle(resp jsonAPISingleResponse) rawTransportTrailerRow {
	attrs := resp.Data.Attributes

	row := rawTransportTrailerRow{
		ID:                resp.Data.ID,
		ExternalTrailerID: stringAttr(attrs, "external-trailer-id"),
		Importer:          stringAttr(attrs, "importer"),
		ImportStatus:      stringAttr(attrs, "import-status"),
	}

	row.BrokerID = relationshipIDFromMap(resp.Data.Relationships, "broker")
	row.TrailerID = relationshipIDFromMap(resp.Data.Relationships, "trailer")

	return row
}
