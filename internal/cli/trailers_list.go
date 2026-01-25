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

type trailersListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Trucker               string
	Tractor               string
	TrailerClassification string
	Broker                string
	InService             string
	Number                string
	NumberLike            string
	Q                     string
	LastShiftStartAtMin   string
	LastShiftStartAtMax   string
}

func newTrailersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trailers",
		Long: `List trailers with filtering and pagination.

Trailers are the cargo-carrying units in a trucking fleet.

Output Columns:
  ID              Trailer identifier
  NUMBER          Trailer number/identifier
  TRUCKER         Trucker ID
  KIND            Trailer kind
  COMPOSITION     Trailer composition
  CAPACITY (LBS)  Capacity in pounds
  IN SERVICE      Whether the trailer is in service

Filters:
  --trucker                   Filter by trucker ID
  --tractor                   Filter by tractor ID
  --trailer-classification    Filter by trailer classification ID
  --broker                    Filter by broker ID
  --in-service                Filter by in-service status (true/false)
  --number                    Filter by trailer number (exact match)
  --number-like               Filter by trailer number (fuzzy match)
  --q                         Search query
  --last-shift-start-at-min   Filter by minimum last shift start time
  --last-shift-start-at-max   Filter by maximum last shift start time`,
		Example: `  # List all trailers
  xbe view trailers list

  # Filter by trucker
  xbe view trailers list --trucker 123

  # Filter by in-service status
  xbe view trailers list --in-service true

  # Search trailers
  xbe view trailers list --q "dump"

  # Filter by trailer number
  xbe view trailers list --number-like "TR100"

  # Output as JSON
  xbe view trailers list --json`,
		RunE: runTrailersList,
	}
	initTrailersListFlags(cmd)
	return cmd
}

func init() {
	trailersCmd.AddCommand(newTrailersListCmd())
}

func initTrailersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("in-service", "", "Filter by in-service status (true/false)")
	cmd.Flags().String("number", "", "Filter by trailer number (exact match)")
	cmd.Flags().String("number-like", "", "Filter by trailer number (fuzzy match)")
	cmd.Flags().String("q", "", "Search query")
	cmd.Flags().String("last-shift-start-at-min", "", "Filter by minimum last shift start time")
	cmd.Flags().String("last-shift-start-at-max", "", "Filter by maximum last shift start time")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTrailersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTrailersListOptions(cmd)
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
	query.Set("fields[trailers]", "number,kind,composition,capacity-lbs,in-service,trucker")
	query.Set("include", "trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[trailer_classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[in_service]", opts.InService)
	setFilterIfPresent(query, "filter[number]", opts.Number)
	setFilterIfPresent(query, "filter[number_like]", opts.NumberLike)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[last_shift_start_at][min]", opts.LastShiftStartAtMin)
	setFilterIfPresent(query, "filter[last_shift_start_at][max]", opts.LastShiftStartAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/trailers", query)
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

	rows := buildTrailerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTrailersTable(cmd, rows)
}

func parseTrailersListOptions(cmd *cobra.Command) (trailersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	trucker, _ := cmd.Flags().GetString("trucker")
	tractor, _ := cmd.Flags().GetString("tractor")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	broker, _ := cmd.Flags().GetString("broker")
	inService, _ := cmd.Flags().GetString("in-service")
	number, _ := cmd.Flags().GetString("number")
	numberLike, _ := cmd.Flags().GetString("number-like")
	q, _ := cmd.Flags().GetString("q")
	lastShiftStartAtMin, _ := cmd.Flags().GetString("last-shift-start-at-min")
	lastShiftStartAtMax, _ := cmd.Flags().GetString("last-shift-start-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return trailersListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Trucker:               trucker,
		Tractor:               tractor,
		TrailerClassification: trailerClassification,
		Broker:                broker,
		InService:             inService,
		Number:                number,
		NumberLike:            numberLike,
		Q:                     q,
		LastShiftStartAtMin:   lastShiftStartAtMin,
		LastShiftStartAtMax:   lastShiftStartAtMax,
	}, nil
}

type trailerRow struct {
	ID          string `json:"id"`
	Number      string `json:"number,omitempty"`
	TruckerID   string `json:"trucker_id,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Composition string `json:"composition,omitempty"`
	CapacityLbs string `json:"capacity_lbs,omitempty"`
	InService   bool   `json:"in_service"`
}

func buildTrailerRows(resp jsonAPIResponse) []trailerRow {
	rows := make([]trailerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := trailerRow{
			ID:          resource.ID,
			Number:      stringAttr(resource.Attributes, "number"),
			Kind:        stringAttr(resource.Attributes, "kind"),
			Composition: stringAttr(resource.Attributes, "composition"),
			CapacityLbs: stringAttr(resource.Attributes, "capacity-lbs"),
			InService:   boolAttr(resource.Attributes, "in-service"),
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTrailersTable(cmd *cobra.Command, rows []trailerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trailers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNUMBER\tTRUCKER\tKIND\tCOMPOSITION\tCAPACITY (LBS)\tIN SERVICE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%v\n",
			row.ID,
			row.Number,
			row.TruckerID,
			row.Kind,
			row.Composition,
			row.CapacityLbs,
			row.InService,
		)
	}
	return writer.Flush()
}
