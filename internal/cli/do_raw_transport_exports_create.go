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

type doRawTransportExportsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ExternalOrderNumber  string
	TargetDatabase       string
	TargetTable          string
	ExportType           string
	Headers              string
	Rows                 string
	FormattedExport      string
	Checksum             string
	Sequence             string
	StpNumbers           string
	IsExportable         bool
	IsExported           bool
	NotExportableReasons string
	IssueType            string
	FirstSeenAt          string
	ThrottledUntil       string
	ExportResults        string
	ExportedAt           string
	Broker               string
	TransportOrder       string
	CreatedBy            string
}

func newDoRawTransportExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw transport export",
		Long: `Create a raw transport export.

Required flags:
  --external-order-number  External order number (required)
  --target-database        Target database (required)

Optional flags:
  --target-table           Target table
  --export-type            Export type
  --headers                Headers payload as JSON
  --rows                   Rows payload as JSON
  --formatted-export       Formatted export payload
  --checksum               Export checksum
  --sequence               Export sequence
  --stp-numbers            Stop numbers as JSON array
  --is-exportable          Exportable flag (true/false)
  --is-exported            Exported flag (true/false)
  --not-exportable-reasons Non-exportable reasons as JSON array
  --issue-type             Issue type
  --first-seen-at          First seen timestamp (ISO 8601)
  --throttled-until        Throttled until timestamp (ISO 8601)
  --export-results         Export results payload as JSON
  --exported-at            Exported timestamp (ISO 8601)
  --broker                 Broker ID
  --transport-order        Transport order ID
  --created-by             Creator user ID`,
		Example: `  # Create a raw transport export
  xbe do raw-transport-exports create \
    --external-order-number ORD-123 \
    --target-database tmw \
    --target-table stops \
    --export-type quantix_tmw \
    --broker 456

  # Create with payloads
  xbe do raw-transport-exports create \
    --external-order-number ORD-456 \
    --target-database tmw \
    --headers '["col1","col2"]' \
    --rows '[["val1","val2"]]'

  # Output as JSON
  xbe do raw-transport-exports create --external-order-number ORD-789 --target-database tmw --json`,
		Args: cobra.NoArgs,
		RunE: runDoRawTransportExportsCreate,
	}
	initDoRawTransportExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportExportsCmd.AddCommand(newDoRawTransportExportsCreateCmd())
}

func initDoRawTransportExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-order-number", "", "External order number (required)")
	cmd.Flags().String("target-database", "", "Target database (required)")
	cmd.Flags().String("target-table", "", "Target table")
	cmd.Flags().String("export-type", "", "Export type")
	cmd.Flags().String("headers", "", "Headers payload as JSON")
	cmd.Flags().String("rows", "", "Rows payload as JSON")
	cmd.Flags().String("formatted-export", "", "Formatted export payload")
	cmd.Flags().String("checksum", "", "Export checksum")
	cmd.Flags().String("sequence", "", "Export sequence")
	cmd.Flags().String("stp-numbers", "", "Stop numbers as JSON array")
	cmd.Flags().Bool("is-exportable", false, "Exportable flag")
	cmd.Flags().Bool("is-exported", false, "Exported flag")
	cmd.Flags().String("not-exportable-reasons", "", "Non-exportable reasons as JSON array")
	cmd.Flags().String("issue-type", "", "Issue type")
	cmd.Flags().String("first-seen-at", "", "First seen timestamp (ISO 8601)")
	cmd.Flags().String("throttled-until", "", "Throttled until timestamp (ISO 8601)")
	cmd.Flags().String("export-results", "", "Export results payload as JSON")
	cmd.Flags().String("exported-at", "", "Exported timestamp (ISO 8601)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("transport-order", "", "Transport order ID")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("external-order-number")
	_ = cmd.MarkFlagRequired("target-database")
}

func runDoRawTransportExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawTransportExportsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ExternalOrderNumber) == "" {
		err := fmt.Errorf("--external-order-number is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TargetDatabase) == "" {
		err := fmt.Errorf("--target-database is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"external-order-number": opts.ExternalOrderNumber,
		"target-database":       opts.TargetDatabase,
	}

	if strings.TrimSpace(opts.TargetTable) != "" {
		attributes["target-table"] = opts.TargetTable
	}
	if strings.TrimSpace(opts.ExportType) != "" {
		attributes["export-type"] = opts.ExportType
	}
	if strings.TrimSpace(opts.FormattedExport) != "" {
		attributes["formatted-export"] = opts.FormattedExport
	}
	if strings.TrimSpace(opts.Checksum) != "" {
		attributes["checksum"] = opts.Checksum
	}
	if strings.TrimSpace(opts.Sequence) != "" {
		attributes["sequence"] = opts.Sequence
	}
	if cmd.Flags().Changed("is-exportable") {
		attributes["is-exportable"] = opts.IsExportable
	}
	if cmd.Flags().Changed("is-exported") {
		attributes["is-exported"] = opts.IsExported
	}
	if strings.TrimSpace(opts.IssueType) != "" {
		attributes["issue-type"] = opts.IssueType
	}
	if strings.TrimSpace(opts.FirstSeenAt) != "" {
		attributes["first-seen-at"] = opts.FirstSeenAt
	}
	if strings.TrimSpace(opts.ThrottledUntil) != "" {
		attributes["throttled-until"] = opts.ThrottledUntil
	}
	if strings.TrimSpace(opts.ExportedAt) != "" {
		attributes["exported-at"] = opts.ExportedAt
	}

	if strings.TrimSpace(opts.Headers) != "" {
		payload, err := parseRawTransportExportJSON("headers", opts.Headers)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["headers"] = payload
	}
	if strings.TrimSpace(opts.Rows) != "" {
		payload, err := parseRawTransportExportJSON("rows", opts.Rows)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["rows"] = payload
	}
	if strings.TrimSpace(opts.ExportResults) != "" {
		payload, err := parseRawTransportExportJSON("export-results", opts.ExportResults)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["export-results"] = payload
	}
	if strings.TrimSpace(opts.StpNumbers) != "" {
		payload, err := parseRawTransportExportJSONArray("stp-numbers", opts.StpNumbers)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["stp-numbers"] = payload
	}
	if strings.TrimSpace(opts.NotExportableReasons) != "" {
		payload, err := parseRawTransportExportJSONArray("not-exportable-reasons", opts.NotExportableReasons)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["not-exportable-reasons"] = payload
	}

	relationships := map[string]any{}
	if strings.TrimSpace(opts.Broker) != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if strings.TrimSpace(opts.TransportOrder) != "" {
		relationships["transport-order"] = map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   opts.TransportOrder,
			},
		}
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "raw-transport-exports",
			"attributes": attributes,
		},
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/raw-transport-exports", jsonBody)
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

	row := rawTransportExportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport export %s\n", row.ID)
	return nil
}

func parseDoRawTransportExportsCreateOptions(cmd *cobra.Command) (doRawTransportExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalOrderNumber, _ := cmd.Flags().GetString("external-order-number")
	targetDatabase, _ := cmd.Flags().GetString("target-database")
	targetTable, _ := cmd.Flags().GetString("target-table")
	exportType, _ := cmd.Flags().GetString("export-type")
	headers, _ := cmd.Flags().GetString("headers")
	rows, _ := cmd.Flags().GetString("rows")
	formattedExport, _ := cmd.Flags().GetString("formatted-export")
	checksum, _ := cmd.Flags().GetString("checksum")
	sequence, _ := cmd.Flags().GetString("sequence")
	stpNumbers, _ := cmd.Flags().GetString("stp-numbers")
	isExportable, _ := cmd.Flags().GetBool("is-exportable")
	isExported, _ := cmd.Flags().GetBool("is-exported")
	notExportableReasons, _ := cmd.Flags().GetString("not-exportable-reasons")
	issueType, _ := cmd.Flags().GetString("issue-type")
	firstSeenAt, _ := cmd.Flags().GetString("first-seen-at")
	throttledUntil, _ := cmd.Flags().GetString("throttled-until")
	exportResults, _ := cmd.Flags().GetString("export-results")
	exportedAt, _ := cmd.Flags().GetString("exported-at")
	broker, _ := cmd.Flags().GetString("broker")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportExportsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ExternalOrderNumber:  externalOrderNumber,
		TargetDatabase:       targetDatabase,
		TargetTable:          targetTable,
		ExportType:           exportType,
		Headers:              headers,
		Rows:                 rows,
		FormattedExport:      formattedExport,
		Checksum:             checksum,
		Sequence:             sequence,
		StpNumbers:           stpNumbers,
		IsExportable:         isExportable,
		IsExported:           isExported,
		NotExportableReasons: notExportableReasons,
		IssueType:            issueType,
		FirstSeenAt:          firstSeenAt,
		ThrottledUntil:       throttledUntil,
		ExportResults:        exportResults,
		ExportedAt:           exportedAt,
		Broker:               broker,
		TransportOrder:       transportOrder,
		CreatedBy:            createdBy,
	}, nil
}

func parseRawTransportExportJSON(label, raw string) (any, error) {
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("invalid %s JSON: %w", label, err)
	}
	return payload, nil
}

func parseRawTransportExportJSONArray(label, raw string) ([]any, error) {
	payload, err := parseRawTransportExportJSON(label, raw)
	if err != nil {
		return nil, err
	}
	items, ok := payload.([]any)
	if !ok {
		return nil, fmt.Errorf("--%s must be a JSON array", label)
	}
	return items, nil
}
