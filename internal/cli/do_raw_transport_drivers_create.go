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

type doRawTransportDriversCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	BrokerID         string
	ExternalDriverID string
	Importer         string
	Tables           string
}

func newDoRawTransportDriversCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw transport driver",
		Long: `Create a raw transport driver record.

Required flags:
  --broker               Broker ID
  --external-driver-id   External driver identifier
  --tables               Tables JSON array (raw transport payload)

Optional flags:
  --importer             Importer key (e.g., quantix_tmw)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a raw transport driver with empty tables
  xbe do raw-transport-drivers create \
    --broker 123 \
    --external-driver-id DRV-0001 \
    --importer quantix_tmw \
    --tables '[]'

  # Output as JSON
  xbe do raw-transport-drivers create --broker 123 --external-driver-id DRV-0002 --tables '[]' --json`,
		Args: cobra.NoArgs,
		RunE: runDoRawTransportDriversCreate,
	}
	initDoRawTransportDriversCreateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportDriversCmd.AddCommand(newDoRawTransportDriversCreateCmd())
}

func initDoRawTransportDriversCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("external-driver-id", "", "External driver identifier (required)")
	cmd.Flags().String("tables", "", "Tables JSON array (required)")
	cmd.Flags().String("importer", "", "Importer key (e.g., quantix_tmw)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("external-driver-id")
	cmd.MarkFlagRequired("tables")
}

func runDoRawTransportDriversCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawTransportDriversCreateOptions(cmd)
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

	externalDriverID := strings.TrimSpace(opts.ExternalDriverID)
	if externalDriverID == "" {
		err := fmt.Errorf("--external-driver-id is required")
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
		"external-driver-id": externalDriverID,
		"tables":             tables,
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
			"type":          "raw-transport-drivers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/raw-transport-drivers", jsonBody)
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

	row := buildRawTransportDriverRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport driver %s\n", row.ID)
	return nil
}

func parseDoRawTransportDriversCreateOptions(cmd *cobra.Command) (doRawTransportDriversCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	externalDriverID, _ := cmd.Flags().GetString("external-driver-id")
	importer, _ := cmd.Flags().GetString("importer")
	tables, _ := cmd.Flags().GetString("tables")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportDriversCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		BrokerID:         brokerID,
		ExternalDriverID: externalDriverID,
		Importer:         importer,
		Tables:           tables,
	}, nil
}

func buildRawTransportDriverRowFromSingle(resp jsonAPISingleResponse) rawTransportDriverRow {
	attrs := resp.Data.Attributes

	row := rawTransportDriverRow{
		ID:               resp.Data.ID,
		ExternalDriverID: stringAttr(attrs, "external-driver-id"),
		Importer:         stringAttr(attrs, "importer"),
		ImportStatus:     stringAttr(attrs, "import-status"),
	}

	row.BrokerID = relationshipIDFromMap(resp.Data.Relationships, "broker")
	row.UserID = relationshipIDFromMap(resp.Data.Relationships, "user")
	row.TruckerMembershipID = relationshipIDFromMap(resp.Data.Relationships, "trucker-membership")

	return row
}
