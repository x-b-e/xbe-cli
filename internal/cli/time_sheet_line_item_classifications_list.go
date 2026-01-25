package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type timeSheetLineItemClassificationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

func newTimeSheetLineItemClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet line item classifications",
		Long: `List time sheet line item classifications with pagination.

Time sheet line item classifications categorize line items on time sheets.

Output Columns:
  ID            Classification identifier
  NAME          Classification name
  DESCRIPTION   Description
  SUBJECT TYPES Types this classification applies to`,
		Example: `  # List all time sheet line item classifications
  xbe view time-sheet-line-item-classifications list

  # Output as JSON
  xbe view time-sheet-line-item-classifications list --json`,
		RunE: runTimeSheetLineItemClassificationsList,
	}
	initTimeSheetLineItemClassificationsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetLineItemClassificationsCmd.AddCommand(newTimeSheetLineItemClassificationsListCmd())
}

func initTimeSheetLineItemClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetLineItemClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetLineItemClassificationsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "name")
	query.Set("fields[time-sheet-line-item-classifications]", "name,description,subject-types")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-line-item-classifications", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildTimeSheetLineItemClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetLineItemClassificationsTable(cmd, rows)
}

func parseTimeSheetLineItemClassificationsListOptions(cmd *cobra.Command) (timeSheetLineItemClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetLineItemClassificationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

type timeSheetLineItemClassificationRow struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	SubjectTypes []string `json:"subject_types,omitempty"`
}

func buildTimeSheetLineItemClassificationRows(resp jsonAPIResponse) []timeSheetLineItemClassificationRow {
	rows := make([]timeSheetLineItemClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := timeSheetLineItemClassificationRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			Description: stringAttr(resource.Attributes, "description"),
		}

		if st, ok := resource.Attributes["subject-types"].([]any); ok {
			for _, s := range st {
				if str, ok := s.(string); ok {
					row.SubjectTypes = append(row.SubjectTypes, str)
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTimeSheetLineItemClassificationsTable(cmd *cobra.Command, rows []timeSheetLineItemClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheet line item classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tSUBJECT TYPES")
	for _, row := range rows {
		subjectTypes := strings.Join(row.SubjectTypes, ", ")
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Description, 30),
			truncateString(subjectTypes, 30),
		)
	}
	return writer.Flush()
}
