package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type placePredictionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Q       string
}

type placePredictionRow struct {
	ID          string `json:"id"`
	PlaceID     string `json:"place_id,omitempty"`
	Description string `json:"description,omitempty"`
}

func newPlacePredictionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List place predictions",
		Long: `List place predictions for a location query.

Place predictions are autocomplete suggestions for a query string.
Results include the Google Place ID and display description.

Output Columns:
  PLACE_ID     Google Place identifier
  DESCRIPTION  Suggested place description

Filters:
  --q  Query string used for place predictions

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # List predictions
  xbe view place-predictions list --q "Austin"

  # Output as JSON
  xbe view place-predictions list --q "Austin" --json`,
		Args: cobra.NoArgs,
		RunE: runPlacePredictionsList,
	}
	initPlacePredictionsListFlags(cmd)
	return cmd
}

func init() {
	placePredictionsCmd.AddCommand(newPlacePredictionsListCmd())
}

func initPlacePredictionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("q", "", "Query string for place predictions")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPlacePredictionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePlacePredictionsListOptions(cmd)
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
	query.Set("fields[place-predictions]", "place-id,description")
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/place-predictions", query)
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

	rows := buildPlacePredictionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPlacePredictionsTable(cmd, rows)
}

func parsePlacePredictionsListOptions(cmd *cobra.Command) (placePredictionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return placePredictionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Q:       q,
	}, nil
}

func buildPlacePredictionRows(resp jsonAPIResponse) []placePredictionRow {
	rows := make([]placePredictionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		placeID := stringAttr(resource.Attributes, "place-id")
		if placeID == "" {
			placeID = resource.ID
		}
		rows = append(rows, placePredictionRow{
			ID:          resource.ID,
			PlaceID:     placeID,
			Description: stringAttr(resource.Attributes, "description"),
		})
	}
	return rows
}

func renderPlacePredictionsTable(cmd *cobra.Command, rows []placePredictionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No place predictions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "PLACE_ID\tDESCRIPTION")
	for _, row := range rows {
		placeID := row.PlaceID
		if placeID == "" {
			placeID = row.ID
		}
		fmt.Fprintf(writer, "%s\t%s\n", placeID, row.Description)
	}
	return writer.Flush()
}
