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

type craftClassesListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Craft             string
	Broker            string
	IsValidForDrivers string
}

func newCraftClassesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List craft classes",
		Long: `List craft classes with filtering and pagination.

Craft classes are sub-classifications within a craft, used to categorize laborers.

Output Columns:
  ID                    Craft class identifier
  NAME                  Craft class name
  CODE                  Craft class code
  VALID FOR DRIVERS     Whether valid for drivers
  CRAFT                 Parent craft name

Filters:
  --craft                Filter by craft ID
  --broker               Filter by broker ID
  --is-valid-for-drivers Filter by driver validity (true/false)`,
		Example: `  # List all craft classes
  xbe view craft-classes list

  # Filter by craft
  xbe view craft-classes list --craft 123

  # Filter by broker
  xbe view craft-classes list --broker 456

  # Show only classes valid for drivers
  xbe view craft-classes list --is-valid-for-drivers true

  # Output as JSON
  xbe view craft-classes list --json`,
		RunE: runCraftClassesList,
	}
	initCraftClassesListFlags(cmd)
	return cmd
}

func init() {
	craftClassesCmd.AddCommand(newCraftClassesListCmd())
}

func initCraftClassesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("craft", "", "Filter by craft ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("is-valid-for-drivers", "", "Filter by driver validity (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCraftClassesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCraftClassesListOptions(cmd)
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
	query.Set("fields[craft-classes]", "name,code,is-valid-for-drivers,craft")
	query.Set("fields[crafts]", "name")
	query.Set("include", "craft")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[craft]", opts.Craft)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[is-valid-for-drivers]", opts.IsValidForDrivers)

	body, _, err := client.Get(cmd.Context(), "/v1/craft-classes", query)
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

	rows := buildCraftClassRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCraftClassesTable(cmd, rows)
}

func parseCraftClassesListOptions(cmd *cobra.Command) (craftClassesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	craft, _ := cmd.Flags().GetString("craft")
	broker, _ := cmd.Flags().GetString("broker")
	isValidForDrivers, _ := cmd.Flags().GetString("is-valid-for-drivers")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return craftClassesListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Craft:             craft,
		Broker:            broker,
		IsValidForDrivers: isValidForDrivers,
	}, nil
}

type craftClassRow struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Code              string `json:"code,omitempty"`
	IsValidForDrivers bool   `json:"is_valid_for_drivers"`
	CraftID           string `json:"craft_id,omitempty"`
	CraftName         string `json:"craft_name,omitempty"`
}

func buildCraftClassRows(resp jsonAPIResponse) []craftClassRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]craftClassRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := craftClassRow{
			ID:                resource.ID,
			Name:              stringAttr(resource.Attributes, "name"),
			Code:              stringAttr(resource.Attributes, "code"),
			IsValidForDrivers: boolAttr(resource.Attributes, "is-valid-for-drivers"),
		}

		if rel, ok := resource.Relationships["craft"]; ok && rel.Data != nil {
			row.CraftID = rel.Data.ID
			if craft, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CraftName = stringAttr(craft.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCraftClassesTable(cmd *cobra.Command, rows []craftClassRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No craft classes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tCODE\tVALID FOR DRIVERS\tCRAFT")
	for _, row := range rows {
		validForDrivers := "No"
		if row.IsValidForDrivers {
			validForDrivers = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Code, 15),
			validForDrivers,
			truncateString(row.CraftName, 25),
		)
	}
	return writer.Flush()
}
