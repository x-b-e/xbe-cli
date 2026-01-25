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

type rawTransportDriversListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	Importer     string
	ImportStatus string
}

type rawTransportDriverRow struct {
	ID                  string `json:"id"`
	ExternalDriverID    string `json:"external_driver_id,omitempty"`
	Importer            string `json:"importer,omitempty"`
	ImportStatus        string `json:"import_status,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	TruckerMembershipID string `json:"trucker_membership_id,omitempty"`
}

func newRawTransportDriversListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw transport drivers",
		Long: `List raw transport drivers with filtering and pagination.

Output Columns:
  ID                 Raw transport driver ID
  EXTERNAL ID        External driver identifier
  IMPORTER           Importer key
  STATUS             Import status
  BROKER             Broker ID
  USER               User ID
  TRUCKER MEMBERSHIP Trucker membership ID

Filters:
  --broker         Filter by broker ID
  --importer       Filter by importer key
  --import-status  Filter by import status (pending, success, failed)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw transport drivers
  xbe view raw-transport-drivers list

  # Filter by broker
  xbe view raw-transport-drivers list --broker 123

  # Filter by importer
  xbe view raw-transport-drivers list --importer quantix_tmw

  # Filter by import status
  xbe view raw-transport-drivers list --import-status pending

  # Output as JSON
  xbe view raw-transport-drivers list --json`,
		Args: cobra.NoArgs,
		RunE: runRawTransportDriversList,
	}
	initRawTransportDriversListFlags(cmd)
	return cmd
}

func init() {
	rawTransportDriversCmd.AddCommand(newRawTransportDriversListCmd())
}

func initRawTransportDriversListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("importer", "", "Filter by importer key")
	cmd.Flags().String("import-status", "", "Filter by import status (pending, success, failed)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportDriversList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawTransportDriversListOptions(cmd)
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
	query.Set("fields[raw-transport-drivers]", "external-driver-id,importer,import-status")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[importer]", opts.Importer)
	setFilterIfPresent(query, "filter[import-status]", opts.ImportStatus)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-drivers", query)
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

	rows := buildRawTransportDriverRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawTransportDriversTable(cmd, rows)
}

func parseRawTransportDriversListOptions(cmd *cobra.Command) (rawTransportDriversListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	importer, _ := cmd.Flags().GetString("importer")
	importStatus, _ := cmd.Flags().GetString("import-status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportDriversListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Broker:       broker,
		Importer:     importer,
		ImportStatus: importStatus,
	}, nil
}

func buildRawTransportDriverRows(resp jsonAPIResponse) []rawTransportDriverRow {
	rows := make([]rawTransportDriverRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := rawTransportDriverRow{
			ID:                  resource.ID,
			ExternalDriverID:    stringAttr(attrs, "external-driver-id"),
			Importer:            stringAttr(attrs, "importer"),
			ImportStatus:        stringAttr(attrs, "import-status"),
			BrokerID:            relationshipIDFromMap(resource.Relationships, "broker"),
			UserID:              relationshipIDFromMap(resource.Relationships, "user"),
			TruckerMembershipID: relationshipIDFromMap(resource.Relationships, "trucker-membership"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderRawTransportDriversTable(cmd *cobra.Command, rows []rawTransportDriverRow) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL ID\tIMPORTER\tSTATUS\tBROKER\tUSER\tTRUCKER MEMBERSHIP")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalDriverID, 24),
			truncateString(row.Importer, 24),
			row.ImportStatus,
			row.BrokerID,
			row.UserID,
			row.TruckerMembershipID,
		)
	}
	return writer.Flush()
}
