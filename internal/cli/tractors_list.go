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

type tractorsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Trucker             string
	Broker              string
	InService           string
	Number              string
	NumberLike          string
	LastShiftStartAtMin string
	LastShiftStartAtMax string
}

func newTractorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tractors",
		Long: `List tractors with filtering and pagination.

Tractors are the power units (trucks) in a trucking fleet.

Output Columns:
  ID            Tractor identifier
  NUMBER        Tractor number/identifier
  TRUCKER       Trucker ID
  MANUFACTURER  Truck manufacturer name
  MODEL         Truck model name
  YEAR          Model year
  IN SERVICE    Whether the tractor is in service

Filters:
  --trucker                     Filter by trucker ID
  --broker                      Filter by broker ID
  --in-service                  Filter by in-service status (true/false)
  --number                      Filter by tractor number (exact match)
  --number-like                 Filter by tractor number (fuzzy match)
  --last-shift-start-at-min     Filter by minimum last shift start time
  --last-shift-start-at-max     Filter by maximum last shift start time`,
		Example: `  # List all tractors
  xbe view tractors list

  # Filter by trucker
  xbe view tractors list --trucker 123

  # Filter by in-service status
  xbe view tractors list --in-service true

  # Filter by tractor number
  xbe view tractors list --number-like "T100"

  # Output as JSON
  xbe view tractors list --json`,
		RunE: runTractorsList,
	}
	initTractorsListFlags(cmd)
	return cmd
}

func init() {
	tractorsCmd.AddCommand(newTractorsListCmd())
}

func initTractorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("in-service", "", "Filter by in-service status (true/false)")
	cmd.Flags().String("number", "", "Filter by tractor number (exact match)")
	cmd.Flags().String("number-like", "", "Filter by tractor number (fuzzy match)")
	cmd.Flags().String("last-shift-start-at-min", "", "Filter by minimum last shift start time")
	cmd.Flags().String("last-shift-start-at-max", "", "Filter by maximum last shift start time")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTractorsListOptions(cmd)
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
	query.Set("fields[tractors]", "number,truck-manufacturer-name,truck-model-name,truck-model-year,in-service,trucker")
	query.Set("include", "trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[in_service]", opts.InService)
	setFilterIfPresent(query, "filter[number]", opts.Number)
	setFilterIfPresent(query, "filter[number_like]", opts.NumberLike)
	setFilterIfPresent(query, "filter[last_shift_start_at][min]", opts.LastShiftStartAtMin)
	setFilterIfPresent(query, "filter[last_shift_start_at][max]", opts.LastShiftStartAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/tractors", query)
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

	rows := buildTractorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTractorsTable(cmd, rows)
}

func parseTractorsListOptions(cmd *cobra.Command) (tractorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	inService, _ := cmd.Flags().GetString("in-service")
	number, _ := cmd.Flags().GetString("number")
	numberLike, _ := cmd.Flags().GetString("number-like")
	lastShiftStartAtMin, _ := cmd.Flags().GetString("last-shift-start-at-min")
	lastShiftStartAtMax, _ := cmd.Flags().GetString("last-shift-start-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Trucker:             trucker,
		Broker:              broker,
		InService:           inService,
		Number:              number,
		NumberLike:          numberLike,
		LastShiftStartAtMin: lastShiftStartAtMin,
		LastShiftStartAtMax: lastShiftStartAtMax,
	}, nil
}

type tractorRow struct {
	ID               string `json:"id"`
	Number           string `json:"number,omitempty"`
	TruckerID        string `json:"trucker_id,omitempty"`
	ManufacturerName string `json:"manufacturer_name,omitempty"`
	ModelName        string `json:"model_name,omitempty"`
	ModelYear        string `json:"model_year,omitempty"`
	InService        bool   `json:"in_service"`
}

func buildTractorRows(resp jsonAPIResponse) []tractorRow {
	rows := make([]tractorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tractorRow{
			ID:               resource.ID,
			Number:           stringAttr(resource.Attributes, "number"),
			ManufacturerName: stringAttr(resource.Attributes, "truck-manufacturer-name"),
			ModelName:        stringAttr(resource.Attributes, "truck-model-name"),
			ModelYear:        stringAttr(resource.Attributes, "truck-model-year"),
			InService:        boolAttr(resource.Attributes, "in-service"),
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTractorsTable(cmd *cobra.Command, rows []tractorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tractors found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNUMBER\tTRUCKER\tMANUFACTURER\tMODEL\tYEAR\tIN SERVICE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%v\n",
			row.ID,
			row.Number,
			row.TruckerID,
			row.ManufacturerName,
			row.ModelName,
			row.ModelYear,
			row.InService,
		)
	}
	return writer.Flush()
}
