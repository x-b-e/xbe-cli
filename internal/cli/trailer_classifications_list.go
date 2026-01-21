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

type trailerClassificationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

type trailerClassificationRow struct {
	ID                        string `json:"id"`
	Name                      string `json:"name"`
	Abbreviation              string `json:"abbreviation,omitempty"`
	RearAxleCount             int    `json:"rear_axle_count,omitempty"`
	CapacityLbs               int    `json:"capacity_lbs,omitempty"`
	IsHeavyEquipmentTransport bool   `json:"is_heavy_equipment_transport"`
}

func newTrailerClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trailer classifications",
		Long: `List trailer classifications with pagination.

Trailer classifications define trailer types with their specifications for
matching trailers to job requirements.

Output Columns:
  ID            Trailer classification identifier
  NAME          Classification name (e.g., End Dump, Belly Dump)
  ABBREVIATION  Short code
  AXLES         Number of rear axles
  CAPACITY      Capacity in pounds
  HEAVY EQUIP   Whether used for heavy equipment transport`,
		Example: `  # List all trailer classifications
  xbe view trailer-classifications list

  # Output as JSON
  xbe view trailer-classifications list --json`,
		RunE: runTrailerClassificationsList,
	}
	initTrailerClassificationsListFlags(cmd)
	return cmd
}

func init() {
	trailerClassificationsCmd.AddCommand(newTrailerClassificationsListCmd())
}

func initTrailerClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTrailerClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTrailerClassificationsListOptions(cmd)
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
	query.Set("fields[trailer-classifications]", "name,abbreviation,rear-axle-count,capacity-lbs,is-heavy-equipment-transport")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/trailer-classifications", query)
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

	rows := buildTrailerClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTrailerClassificationsTable(cmd, rows)
}

func parseTrailerClassificationsListOptions(cmd *cobra.Command) (trailerClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return trailerClassificationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func buildTrailerClassificationRows(resp jsonAPIResponse) []trailerClassificationRow {
	rows := make([]trailerClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := trailerClassificationRow{
			ID:                        resource.ID,
			Name:                      stringAttr(resource.Attributes, "name"),
			Abbreviation:              stringAttr(resource.Attributes, "abbreviation"),
			RearAxleCount:             intAttr(resource.Attributes, "rear-axle-count"),
			CapacityLbs:               intAttr(resource.Attributes, "capacity-lbs"),
			IsHeavyEquipmentTransport: boolAttr(resource.Attributes, "is-heavy-equipment-transport"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTrailerClassificationsTable(cmd *cobra.Command, rows []trailerClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trailer classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tAXLES\tCAPACITY\tHEAVY EQUIP")
	for _, row := range rows {
		heavyEquip := "no"
		if row.IsHeavyEquipmentTransport {
			heavyEquip = "yes"
		}
		capacity := ""
		if row.CapacityLbs > 0 {
			capacity = fmt.Sprintf("%d lbs", row.CapacityLbs)
		}
		axles := ""
		if row.RearAxleCount > 0 {
			axles = strconv.Itoa(row.RearAxleCount)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Abbreviation, 10),
			axles,
			capacity,
			heavyEquip,
		)
	}
	return writer.Flush()
}
