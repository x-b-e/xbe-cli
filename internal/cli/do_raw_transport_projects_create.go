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

type doRawTransportProjectsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	BrokerID              string
	ExternalProjectNumber string
	Importer              string
	Tables                string
	TablesRowversionMin   string
	TablesRowversionMax   string
	IsManaged             bool
}

func newDoRawTransportProjectsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw transport project",
		Long: `Create a raw transport project record.

Required flags:
  --broker                   Broker ID
  --external-project-number  External project number
  --tables                   Tables JSON array (raw transport payload)

Optional flags:
  --importer                 Importer key (e.g., quantix_tmw)
  --tables-rowversion-min    Tables rowversion minimum
  --tables-rowversion-max    Tables rowversion maximum
  --is-managed               Mark as managed

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a raw transport project with empty tables
  xbe do raw-transport-projects create \
    --broker 123 \
    --external-project-number PROJ-0001 \
    --importer quantix_tmw \
    --tables '[]'

  # Output as JSON
  xbe do raw-transport-projects create --broker 123 --external-project-number PROJ-0002 --tables '[]' --json`,
		Args: cobra.NoArgs,
		RunE: runDoRawTransportProjectsCreate,
	}
	initDoRawTransportProjectsCreateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportProjectsCmd.AddCommand(newDoRawTransportProjectsCreateCmd())
}

func initDoRawTransportProjectsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("external-project-number", "", "External project number (required)")
	cmd.Flags().String("tables", "", "Tables JSON array (required)")
	cmd.Flags().String("importer", "", "Importer key (e.g., quantix_tmw)")
	cmd.Flags().String("tables-rowversion-min", "", "Tables rowversion minimum")
	cmd.Flags().String("tables-rowversion-max", "", "Tables rowversion maximum")
	cmd.Flags().Bool("is-managed", false, "Mark as managed")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("external-project-number")
	cmd.MarkFlagRequired("tables")
}

func runDoRawTransportProjectsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawTransportProjectsCreateOptions(cmd)
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

	externalProjectNumber := strings.TrimSpace(opts.ExternalProjectNumber)
	if externalProjectNumber == "" {
		err := fmt.Errorf("--external-project-number is required")
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
		"external-project-number": externalProjectNumber,
		"tables":                  tables,
	}
	if strings.TrimSpace(opts.Importer) != "" {
		attributes["importer"] = strings.TrimSpace(opts.Importer)
	}
	if strings.TrimSpace(opts.TablesRowversionMin) != "" {
		attributes["tables-rowversion-min"] = strings.TrimSpace(opts.TablesRowversionMin)
	}
	if strings.TrimSpace(opts.TablesRowversionMax) != "" {
		attributes["tables-rowversion-max"] = strings.TrimSpace(opts.TablesRowversionMax)
	}
	if cmd.Flags().Changed("is-managed") {
		attributes["is-managed"] = opts.IsManaged
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
			"type":          "raw-transport-projects",
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

	body, _, err := client.Post(cmd.Context(), "/v1/raw-transport-projects", jsonBody)
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

	row := buildRawTransportProjectRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport project %s\n", row.ID)
	return nil
}

func parseDoRawTransportProjectsCreateOptions(cmd *cobra.Command) (doRawTransportProjectsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	externalProjectNumber, _ := cmd.Flags().GetString("external-project-number")
	importer, _ := cmd.Flags().GetString("importer")
	tables, _ := cmd.Flags().GetString("tables")
	rowversionMin, _ := cmd.Flags().GetString("tables-rowversion-min")
	rowversionMax, _ := cmd.Flags().GetString("tables-rowversion-max")
	isManaged, _ := cmd.Flags().GetBool("is-managed")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportProjectsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		BrokerID:              brokerID,
		ExternalProjectNumber: externalProjectNumber,
		Importer:              importer,
		Tables:                tables,
		TablesRowversionMin:   rowversionMin,
		TablesRowversionMax:   rowversionMax,
		IsManaged:             isManaged,
	}, nil
}

func buildRawTransportProjectRowFromSingle(resp jsonAPISingleResponse) rawTransportProjectRow {
	attrs := resp.Data.Attributes

	row := rawTransportProjectRow{
		ID:                    resp.Data.ID,
		ExternalProjectNumber: stringAttr(attrs, "external-project-number"),
		Importer:              stringAttr(attrs, "importer"),
		ImportStatus:          stringAttr(attrs, "import-status"),
		IsManaged:             boolAttr(attrs, "is-managed"),
	}

	row.BrokerID = relationshipIDFromMap(resp.Data.Relationships, "broker")
	row.ProjectID = relationshipIDFromMap(resp.Data.Relationships, "project")

	return row
}
