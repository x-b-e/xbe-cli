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

type transportReferencesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	SubjectType string
	SubjectID   string
	Key         string
	Position    string
}

type transportReferenceRow struct {
	ID          string `json:"id"`
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Position    int    `json:"position,omitempty"`
	SubjectType string `json:"subject_type,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
}

func newTransportReferencesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transport references",
		Long: `List transport references with filtering and pagination.

Output Columns:
  ID        Transport reference identifier
  KEY       Reference key
  VALUE     Reference value
  POSITION  Reference position within the subject
  SUBJECT   Subject type and ID

Filters:
  --subject-type  Filter by subject type (e.g., transport-orders, TransportOrder)
  --subject-id    Filter by subject ID
  --key           Filter by reference key
  --position      Filter by reference position

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List transport references
  xbe view transport-references list

  # Filter by subject
  xbe view transport-references list --subject-type transport-orders --subject-id 123

  # Filter by key
  xbe view transport-references list --key BOL

  # Output as JSON
  xbe view transport-references list --json`,
		RunE: runTransportReferencesList,
	}
	initTransportReferencesListFlags(cmd)
	return cmd
}

func init() {
	transportReferencesCmd.AddCommand(newTransportReferencesListCmd())
}

func initTransportReferencesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("subject-type", "", "Filter by subject type (e.g., transport-orders, TransportOrder)")
	cmd.Flags().String("subject-id", "", "Filter by subject ID")
	cmd.Flags().String("key", "", "Filter by reference key")
	cmd.Flags().String("position", "", "Filter by reference position")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportReferencesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTransportReferencesListOptions(cmd)
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
	query.Set("fields[transport-references]", "key,value,position,subject")
	query.Set("include", "subject")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "id")
	}

	setFilterIfPresent(query, "filter[key]", opts.Key)
	setFilterIfPresent(query, "filter[position]", opts.Position)

	if opts.SubjectType != "" && opts.SubjectID != "" {
		subjectType := normalizePolymorphicFilterType(opts.SubjectType)
		query.Set("filter[subject]", subjectType+"|"+opts.SubjectID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/transport-references", query)
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

	rows := buildTransportReferenceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTransportReferencesTable(cmd, rows)
}

func parseTransportReferencesListOptions(cmd *cobra.Command) (transportReferencesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	subjectType, err := cmd.Flags().GetString("subject-type")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	subjectID, err := cmd.Flags().GetString("subject-id")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	position, err := cmd.Flags().GetString("position")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return transportReferencesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return transportReferencesListOptions{}, err
	}

	return transportReferencesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Key:         key,
		Position:    position,
	}, nil
}

func buildTransportReferenceRows(resp jsonAPIResponse) []transportReferenceRow {
	rows := make([]transportReferenceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := transportReferenceRow{
			ID:       resource.ID,
			Key:      stringAttr(resource.Attributes, "key"),
			Value:    stringAttr(resource.Attributes, "value"),
			Position: intAttr(resource.Attributes, "position"),
		}

		if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildTransportReferenceRowFromSingle(resp jsonAPISingleResponse) transportReferenceRow {
	attrs := resp.Data.Attributes

	row := transportReferenceRow{
		ID:       resp.Data.ID,
		Key:      stringAttr(attrs, "key"),
		Value:    stringAttr(attrs, "value"),
		Position: intAttr(attrs, "position"),
	}

	if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
		row.SubjectType = rel.Data.Type
		row.SubjectID = rel.Data.ID
	}

	return row
}

func renderTransportReferencesTable(cmd *cobra.Command, rows []transportReferenceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport references found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tKEY\tVALUE\tPOSITION\tSUBJECT")
	for _, row := range rows {
		subject := ""
		if row.SubjectType != "" && row.SubjectID != "" {
			subject = row.SubjectType + "/" + row.SubjectID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\n",
			row.ID,
			truncateString(row.Key, 20),
			truncateString(row.Value, 30),
			row.Position,
			truncateString(subject, 30),
		)
	}
	return writer.Flush()
}

func normalizePolymorphicFilterType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed
	}
	if !strings.ContainsAny(trimmed, "-_") {
		return trimmed
	}

	normalized := strings.ReplaceAll(trimmed, "_", "-")
	if strings.HasSuffix(normalized, "s") && len(normalized) > 1 {
		normalized = strings.TrimSuffix(normalized, "s")
	}

	parts := strings.Split(normalized, "-")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if len(lower) == 1 {
			parts[i] = strings.ToUpper(lower)
			continue
		}
		parts[i] = strings.ToUpper(lower[:1]) + lower[1:]
	}

	return strings.Join(parts, "")
}
