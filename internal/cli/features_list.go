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

type featuresListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	PDCAStage           string
	DifferentiationDegree string
	Scale               string
	ReleasedOnMin       string
	ReleasedOnMax       string
}

func newFeaturesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List features",
		Long: `List features with filtering and pagination.

Returns a list of features matching the specified criteria, sorted by release
date (newest first).

Output Columns (table format):
  ID                 Unique feature identifier
  NAME               Feature name (branded or generic)
  PDCA               PDCA stage (plan/do/check/act)
  SCALE              Feature scale

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

PDCA Stages:
  plan, do, check, act`,
		Example: `  # List recent features
  xbe view features list

  # Filter by PDCA stage
  xbe view features list --pdca-stage plan
  xbe view features list --pdca-stage do

  # Filter by scale
  xbe view features list --scale 3

  # Filter by date range
  xbe view features list --released-on-min 2024-01-01 --released-on-max 2024-06-30

  # Paginate results
  xbe view features list --limit 20 --offset 40

  # Output as JSON for scripting
  xbe view features list --json`,
		RunE: runFeaturesList,
	}
	initFeaturesListFlags(cmd)
	return cmd
}

func init() {
	featuresCmd.AddCommand(newFeaturesListCmd())
}

func initFeaturesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("pdca-stage", "", "Filter by PDCA stage (plan/do/check/act)")
	cmd.Flags().String("differentiation-degree", "", "Filter by differentiation degree")
	cmd.Flags().String("scale", "", "Filter by scale")
	cmd.Flags().String("released-on-min", "", "Filter to features released on or after this date (YYYY-MM-DD)")
	cmd.Flags().String("released-on-max", "", "Filter to features released on or before this date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFeaturesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseFeaturesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "-released-on")
	query.Set("fields[features]", "name-generic,name-branded,released-on,pdca-stage,scale,differentiation-degree")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[pdca-stage]", opts.PDCAStage)
	setFilterIfPresent(query, "filter[differentiation-degree]", opts.DifferentiationDegree)
	setFilterIfPresent(query, "filter[scale]", opts.Scale)
	setFilterIfPresent(query, "filter[released-on-min]", opts.ReleasedOnMin)
	setFilterIfPresent(query, "filter[released-on-max]", opts.ReleasedOnMax)

	body, _, err := client.Get(cmd.Context(), "/v1/features", query)
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

	if opts.JSON {
		rows := buildFeatureRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderFeaturesTable(cmd, resp)
}

func parseFeaturesListOptions(cmd *cobra.Command) (featuresListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return featuresListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return featuresListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return featuresListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return featuresListOptions{}, err
	}
	pdcaStage, err := cmd.Flags().GetString("pdca-stage")
	if err != nil {
		return featuresListOptions{}, err
	}
	differentiationDegree, err := cmd.Flags().GetString("differentiation-degree")
	if err != nil {
		return featuresListOptions{}, err
	}
	scale, err := cmd.Flags().GetString("scale")
	if err != nil {
		return featuresListOptions{}, err
	}
	releasedOnMin, err := cmd.Flags().GetString("released-on-min")
	if err != nil {
		return featuresListOptions{}, err
	}
	releasedOnMax, err := cmd.Flags().GetString("released-on-max")
	if err != nil {
		return featuresListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return featuresListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return featuresListOptions{}, err
	}

	return featuresListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		PDCAStage:           pdcaStage,
		DifferentiationDegree: differentiationDegree,
		Scale:               scale,
		ReleasedOnMin:       releasedOnMin,
		ReleasedOnMax:       releasedOnMax,
	}, nil
}

func renderFeaturesTable(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildFeatureRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No features found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tNAME\tPDCA\tSCALE")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", row.ID, row.Name, row.PDCAStage, row.Scale)
	}

	return writer.Flush()
}

type featureRow struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	NameGeneric           string `json:"name_generic"`
	NameBranded           string `json:"name_branded"`
	Released              string `json:"released"`
	PDCAStage             string `json:"pdca_stage"`
	Scale                 string `json:"scale"`
	DifferentiationDegree string `json:"differentiation_degree"`
}

func buildFeatureRows(resp jsonAPIResponse) []featureRow {
	rows := make([]featureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		nameGeneric := strings.TrimSpace(stringAttr(resource.Attributes, "name-generic"))
		nameBranded := strings.TrimSpace(stringAttr(resource.Attributes, "name-branded"))
		name := firstNonEmpty(nameBranded, nameGeneric)

		rows = append(rows, featureRow{
			ID:                    resource.ID,
			Name:                  name,
			NameGeneric:           nameGeneric,
			NameBranded:           nameBranded,
			Released:              formatDate(stringAttr(resource.Attributes, "released-on")),
			PDCAStage:             stringAttr(resource.Attributes, "pdca-stage"),
			Scale:                 stringAttr(resource.Attributes, "scale"),
			DifferentiationDegree: stringAttr(resource.Attributes, "differentiation-degree"),
		})
	}

	return rows
}
