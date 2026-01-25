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

type oneStepGpsVehiclesListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	Broker                string
	Trucker               string
	Trailer               string
	Tractor               string
	HasTrailer            string
	HasTractor            string
	AssignedAtMin         string
	IntegrationIdentifier string
	TrailerSetAtMin       string
	TrailerSetAtMax       string
	IsTrailerSetAt        string
	TractorSetAtMin       string
	TractorSetAtMax       string
	IsTractorSetAt        string
	CreatedAtMin          string
	CreatedAtMax          string
	IsCreatedAt           string
	UpdatedAtMin          string
	UpdatedAtMax          string
	IsUpdatedAt           string
}

type oneStepGpsVehicleRow struct {
	ID                                       string `json:"id"`
	VehicleID                                string `json:"vehicle_id,omitempty"`
	VehicleNumber                            string `json:"vehicle_number,omitempty"`
	IntegrationIdentifier                    string `json:"integration_identifier,omitempty"`
	TrailerSetAt                             string `json:"trailer_set_at,omitempty"`
	TractorSetAt                             string `json:"tractor_set_at,omitempty"`
	SkipTrailerIsNotAlreadyMatchedValidation bool   `json:"skip_trailer_is_not_already_matched_validation,omitempty"`
	SkipTractorIsNotAlreadyMatchedValidation bool   `json:"skip_tractor_is_not_already_matched_validation,omitempty"`
	BrokerID                                 string `json:"broker_id,omitempty"`
	TruckerID                                string `json:"trucker_id,omitempty"`
	TrailerID                                string `json:"trailer_id,omitempty"`
	TractorID                                string `json:"tractor_id,omitempty"`
}

func newOneStepGpsVehiclesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List One Step GPS vehicles",
		Long: `List One Step GPS vehicles with filtering and pagination.

Output Columns:
  ID               One Step GPS vehicle identifier
  VEHICLE NUMBER   One Step GPS vehicle number
  VEHICLE ID       One Step GPS vehicle external ID
  TRUCKER          Trucker ID
  TRAILER          Trailer ID
  TRACTOR          Tractor ID
  INTEGRATION ID   Integration identifier

Filters:
  --broker                 Filter by broker ID
  --trucker                Filter by trucker ID
  --trailer                Filter by trailer ID
  --tractor                Filter by tractor ID
  --has-trailer            Filter by presence of trailer (true/false)
  --has-tractor            Filter by presence of tractor (true/false)
  --assigned-at-min        Filter by minimum assignment time (ISO 8601)
  --integration-identifier Filter by integration identifier
  --trailer-set-at-min     Filter by trailer set on/after (ISO 8601)
  --trailer-set-at-max     Filter by trailer set on/before (ISO 8601)
  --is-trailer-set-at      Filter by presence of trailer-set-at (true/false)
  --tractor-set-at-min     Filter by tractor set on/after (ISO 8601)
  --tractor-set-at-max     Filter by tractor set on/before (ISO 8601)
  --is-tractor-set-at      Filter by presence of tractor-set-at (true/false)
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --is-created-at          Filter by presence of created-at (true/false)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-updated-at          Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List One Step GPS vehicles
  xbe view one-step-gps-vehicles list

  # Filter by broker
  xbe view one-step-gps-vehicles list --broker 123

  # Filter by assignment time
  xbe view one-step-gps-vehicles list --assigned-at-min "2024-01-01T00:00:00Z"

  # Output as JSON
  xbe view one-step-gps-vehicles list --json`,
		Args: cobra.NoArgs,
		RunE: runOneStepGpsVehiclesList,
	}
	initOneStepGpsVehiclesListFlags(cmd)
	return cmd
}

func init() {
	oneStepGpsVehiclesCmd.AddCommand(newOneStepGpsVehiclesListCmd())
}

func initOneStepGpsVehiclesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("has-trailer", "", "Filter by presence of trailer (true/false)")
	cmd.Flags().String("has-tractor", "", "Filter by presence of tractor (true/false)")
	cmd.Flags().String("assigned-at-min", "", "Filter by minimum assignment time (ISO 8601)")
	cmd.Flags().String("integration-identifier", "", "Filter by integration identifier")
	cmd.Flags().String("trailer-set-at-min", "", "Filter by trailer set on/after (ISO 8601)")
	cmd.Flags().String("trailer-set-at-max", "", "Filter by trailer set on/before (ISO 8601)")
	cmd.Flags().String("is-trailer-set-at", "", "Filter by presence of trailer-set-at (true/false)")
	cmd.Flags().String("tractor-set-at-min", "", "Filter by tractor set on/after (ISO 8601)")
	cmd.Flags().String("tractor-set-at-max", "", "Filter by tractor set on/before (ISO 8601)")
	cmd.Flags().String("is-tractor-set-at", "", "Filter by presence of tractor-set-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOneStepGpsVehiclesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOneStepGpsVehiclesListOptions(cmd)
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
	query.Set("fields[one-step-gps-vehicles]", "vehicle-id,vehicle-number,integration-identifier,trailer-set-at,tractor-set-at,skip-trailer-is-not-already-matched-validation,skip-tractor-is-not-already-matched-validation,broker,trucker,trailer,tractor")
	query.Set("include", "broker,trucker,trailer,tractor")

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
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[has-trailer]", opts.HasTrailer)
	setFilterIfPresent(query, "filter[has-tractor]", opts.HasTractor)
	setFilterIfPresent(query, "filter[assigned-at-min]", opts.AssignedAtMin)
	setFilterIfPresent(query, "filter[integration-identifier]", opts.IntegrationIdentifier)
	setFilterIfPresent(query, "filter[trailer-set-at-min]", opts.TrailerSetAtMin)
	setFilterIfPresent(query, "filter[trailer-set-at-max]", opts.TrailerSetAtMax)
	setFilterIfPresent(query, "filter[is-trailer-set-at]", opts.IsTrailerSetAt)
	setFilterIfPresent(query, "filter[tractor-set-at-min]", opts.TractorSetAtMin)
	setFilterIfPresent(query, "filter[tractor-set-at-max]", opts.TractorSetAtMax)
	setFilterIfPresent(query, "filter[is-tractor-set-at]", opts.IsTractorSetAt)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/one-step-gps-vehicles", query)
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

	rows := make([]oneStepGpsVehicleRow, 0, len(resp.Data))
	for _, item := range resp.Data {
		attrs := item.Attributes
		row := oneStepGpsVehicleRow{
			ID:                                       item.ID,
			VehicleID:                                stringAttr(attrs, "vehicle-id"),
			VehicleNumber:                            stringAttr(attrs, "vehicle-number"),
			IntegrationIdentifier:                    stringAttr(attrs, "integration-identifier"),
			TrailerSetAt:                             formatDateTime(stringAttr(attrs, "trailer-set-at")),
			TractorSetAt:                             formatDateTime(stringAttr(attrs, "tractor-set-at")),
			SkipTrailerIsNotAlreadyMatchedValidation: boolAttr(attrs, "skip-trailer-is-not-already-matched-validation"),
			SkipTractorIsNotAlreadyMatchedValidation: boolAttr(attrs, "skip-tractor-is-not-already-matched-validation"),
			BrokerID:                                 relationshipIDFromMap(item.Relationships, "broker"),
			TruckerID:                                relationshipIDFromMap(item.Relationships, "trucker"),
			TrailerID:                                relationshipIDFromMap(item.Relationships, "trailer"),
			TractorID:                                relationshipIDFromMap(item.Relationships, "tractor"),
		}
		rows = append(rows, row)
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No One Step GPS vehicles found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tVEHICLE NUMBER\tVEHICLE ID\tTRUCKER\tTRAILER\tTRACTOR\tINTEGRATION ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.VehicleNumber,
			row.VehicleID,
			row.TruckerID,
			row.TrailerID,
			row.TractorID,
			row.IntegrationIdentifier,
		)
	}
	writer.Flush()

	return nil
}

func parseOneStepGpsVehiclesListOptions(cmd *cobra.Command) (oneStepGpsVehiclesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	trucker, err := cmd.Flags().GetString("trucker")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	trailer, err := cmd.Flags().GetString("trailer")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	tractor, err := cmd.Flags().GetString("tractor")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	hasTrailer, err := cmd.Flags().GetString("has-trailer")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	hasTractor, err := cmd.Flags().GetString("has-tractor")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	assignedAtMin, err := cmd.Flags().GetString("assigned-at-min")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	integrationIdentifier, err := cmd.Flags().GetString("integration-identifier")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	trailerSetAtMin, err := cmd.Flags().GetString("trailer-set-at-min")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	trailerSetAtMax, err := cmd.Flags().GetString("trailer-set-at-max")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	isTrailerSetAt, err := cmd.Flags().GetString("is-trailer-set-at")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	tractorSetAtMin, err := cmd.Flags().GetString("tractor-set-at-min")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	tractorSetAtMax, err := cmd.Flags().GetString("tractor-set-at-max")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	isTractorSetAt, err := cmd.Flags().GetString("is-tractor-set-at")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	createdAtMin, err := cmd.Flags().GetString("created-at-min")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	createdAtMax, err := cmd.Flags().GetString("created-at-max")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	isCreatedAt, err := cmd.Flags().GetString("is-created-at")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	updatedAtMin, err := cmd.Flags().GetString("updated-at-min")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	updatedAtMax, err := cmd.Flags().GetString("updated-at-max")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	isUpdatedAt, err := cmd.Flags().GetString("is-updated-at")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return oneStepGpsVehiclesListOptions{}, err
	}

	return oneStepGpsVehiclesListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		Broker:                broker,
		Trucker:               trucker,
		Trailer:               trailer,
		Tractor:               tractor,
		HasTrailer:            hasTrailer,
		HasTractor:            hasTractor,
		AssignedAtMin:         assignedAtMin,
		IntegrationIdentifier: integrationIdentifier,
		TrailerSetAtMin:       trailerSetAtMin,
		TrailerSetAtMax:       trailerSetAtMax,
		IsTrailerSetAt:        isTrailerSetAt,
		TractorSetAtMin:       tractorSetAtMin,
		TractorSetAtMax:       tractorSetAtMax,
		IsTractorSetAt:        isTractorSetAt,
		CreatedAtMin:          createdAtMin,
		CreatedAtMax:          createdAtMax,
		IsCreatedAt:           isCreatedAt,
		UpdatedAtMin:          updatedAtMin,
		UpdatedAtMax:          updatedAtMax,
		IsUpdatedAt:           isUpdatedAt,
	}, nil
}
