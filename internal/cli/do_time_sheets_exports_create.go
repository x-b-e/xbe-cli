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

type doTimeSheetsExportsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	OrganizationFormatter string
	TimeSheetIDs          []string
}

type timeSheetsExportCreateRow struct {
	ID                      string   `json:"id"`
	Status                  string   `json:"status,omitempty"`
	FileName                string   `json:"file_name,omitempty"`
	MimeType                string   `json:"mime_type,omitempty"`
	OrganizationFormatterID string   `json:"organization_formatter_id,omitempty"`
	TimeSheetIDs            []string `json:"time_sheet_ids,omitempty"`
	BrokerID                string   `json:"broker_id,omitempty"`
	OrganizationType        string   `json:"organization_type,omitempty"`
	OrganizationID          string   `json:"organization_id,omitempty"`
}

func newDoTimeSheetsExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time sheets export",
		Long: `Create a time sheets export.

Required flags:
  --organization-formatter  Organization formatter ID (required)
  --time-sheet-ids          Time sheet IDs (required, comma-separated or repeated)`,
		Example: `  # Create a time sheets export
  xbe do time-sheets-exports create \
    --organization-formatter 123 \
    --time-sheet-ids 456,789

  # Output as JSON
  xbe do time-sheets-exports create \
    --organization-formatter 123 \
    --time-sheet-ids 456,789 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetsExportsCreate,
	}
	initDoTimeSheetsExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetsExportsCmd.AddCommand(newDoTimeSheetsExportsCreateCmd())
}

func initDoTimeSheetsExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-formatter", "", "Organization formatter ID (required)")
	cmd.Flags().StringSlice("time-sheet-ids", nil, "Time sheet IDs (required, comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetsExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetsExportsCreateOptions(cmd)
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

	opts.OrganizationFormatter = strings.TrimSpace(opts.OrganizationFormatter)
	if opts.OrganizationFormatter == "" {
		err := fmt.Errorf("--organization-formatter is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	timeSheetIDs := normalizeIDList(opts.TimeSheetIDs)
	if len(timeSheetIDs) == 0 {
		err := fmt.Errorf("--time-sheet-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"organization-formatter": map[string]any{
			"data": map[string]any{
				"type": "organization-formatters",
				"id":   opts.OrganizationFormatter,
			},
		},
		"time-sheets": map[string]any{
			"data": buildRelationshipDataList(timeSheetIDs, "time-sheets"),
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-sheets-exports",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheets-exports", jsonBody)
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

	row := timeSheetsExportCreateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheets export %s\n", row.ID)
	return nil
}

func timeSheetsExportCreateRowFromSingle(resp jsonAPISingleResponse) timeSheetsExportCreateRow {
	attrs := resp.Data.Attributes
	row := timeSheetsExportCreateRow{
		ID:       resp.Data.ID,
		Status:   stringAttr(attrs, "status"),
		FileName: stringAttr(attrs, "file-name"),
		MimeType: stringAttr(attrs, "mime-type"),
	}

	if rel, ok := resp.Data.Relationships["organization-formatter"]; ok && rel.Data != nil {
		row.OrganizationFormatterID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["time-sheets"]; ok {
		row.TimeSheetIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}

func parseDoTimeSheetsExportsCreateOptions(cmd *cobra.Command) (doTimeSheetsExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationFormatter, _ := cmd.Flags().GetString("organization-formatter")
	timeSheetIDs, _ := cmd.Flags().GetStringSlice("time-sheet-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetsExportsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		OrganizationFormatter: organizationFormatter,
		TimeSheetIDs:          timeSheetIDs,
	}, nil
}

func normalizeIDList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
	}
	return cleaned
}
