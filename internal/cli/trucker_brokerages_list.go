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

type truckerBrokeragesListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	Trucker         string
	BrokeredTrucker string
	CreatedAtMin    string
	CreatedAtMax    string
	UpdatedAtMin    string
	UpdatedAtMax    string
}

type truckerBrokerageRow struct {
	ID                string `json:"id"`
	TruckerID         string `json:"trucker_id,omitempty"`
	BrokeredTruckerID string `json:"brokered_trucker_id,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

func newTruckerBrokeragesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker brokerages",
		Long: `List trucker brokerages.

Output Columns:
  ID               Brokerage identifier
  TRUCKER          Brokering trucker ID
  BROKERED_TRUCKER Brokered trucker ID
  CREATED_AT       Created timestamp

Filters:
  --trucker            Filter by brokering trucker ID
  --brokered-trucker   Filter by brokered trucker ID
  --created-at-min     Filter by created-at on/after (ISO 8601)
  --created-at-max     Filter by created-at on/before (ISO 8601)
  --updated-at-min     Filter by updated-at on/after (ISO 8601)
  --updated-at-max     Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trucker brokerages
  xbe view trucker-brokerages list

  # Filter by brokering trucker
  xbe view trucker-brokerages list --trucker 123

  # Filter by brokered trucker
  xbe view trucker-brokerages list --brokered-trucker 456

  # Output as JSON
  xbe view trucker-brokerages list --json`,
		Args: cobra.NoArgs,
		RunE: runTruckerBrokeragesList,
	}
	initTruckerBrokeragesListFlags(cmd)
	return cmd
}

func init() {
	truckerBrokeragesCmd.AddCommand(newTruckerBrokeragesListCmd())
}

func initTruckerBrokeragesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("trucker", "", "Filter by brokering trucker ID")
	cmd.Flags().String("brokered-trucker", "", "Filter by brokered trucker ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerBrokeragesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerBrokeragesListOptions(cmd)
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
	query.Set("fields[trucker-brokerages]", "created-at,updated-at,trucker,brokered-trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[brokered-trucker]", opts.BrokeredTrucker)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-brokerages", query)
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

	rows := buildTruckerBrokerageRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerBrokeragesTable(cmd, rows)
}

func parseTruckerBrokeragesListOptions(cmd *cobra.Command) (truckerBrokeragesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trucker, _ := cmd.Flags().GetString("trucker")
	brokeredTrucker, _ := cmd.Flags().GetString("brokered-trucker")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerBrokeragesListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		Trucker:         trucker,
		BrokeredTrucker: brokeredTrucker,
		CreatedAtMin:    createdAtMin,
		CreatedAtMax:    createdAtMax,
		UpdatedAtMin:    updatedAtMin,
		UpdatedAtMax:    updatedAtMax,
	}, nil
}

func buildTruckerBrokerageRows(resp jsonAPIResponse) []truckerBrokerageRow {
	rows := make([]truckerBrokerageRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := truckerBrokerageRow{
			ID:                resource.ID,
			TruckerID:         relationshipIDFromMap(resource.Relationships, "trucker"),
			BrokeredTruckerID: relationshipIDFromMap(resource.Relationships, "brokered-trucker"),
			CreatedAt:         formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt:         formatDateTime(stringAttr(attrs, "updated-at")),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTruckerBrokeragesTable(cmd *cobra.Command, rows []truckerBrokerageRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker brokerages found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCKER\tBROKERED_TRUCKER\tCREATED_AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.TruckerID,
			row.BrokeredTruckerID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
