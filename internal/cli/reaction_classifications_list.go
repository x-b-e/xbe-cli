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

type reactionClassificationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

func newReactionClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List reaction classifications",
		Long: `List reaction classifications.

Reaction classifications define the available emoji reactions that can be
used on posts, comments, and other content.

Note: Reaction classifications are read-only and cannot be created,
updated, or deleted through the API.

Output Columns:
  ID                 Classification identifier
  LABEL              Reaction label/name
  UTF8               Emoji character
  EXTERNAL REFERENCE External reference identifier`,
		Example: `  # List all reaction classifications
  xbe view reaction-classifications list

  # Output as JSON
  xbe view reaction-classifications list --json`,
		RunE: runReactionClassificationsList,
	}
	initReactionClassificationsListFlags(cmd)
	return cmd
}

func init() {
	reactionClassificationsCmd.AddCommand(newReactionClassificationsListCmd())
}

func initReactionClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runReactionClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseReactionClassificationsListOptions(cmd)
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
	query.Set("sort", "label")
	query.Set("fields[reaction-classifications]", "label,utf8,external-reference")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/reaction-classifications", query)
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

	rows := buildReactionClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderReactionClassificationsTable(cmd, rows)
}

func parseReactionClassificationsListOptions(cmd *cobra.Command) (reactionClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return reactionClassificationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

type reactionClassificationRow struct {
	ID                string `json:"id"`
	Label             string `json:"label"`
	UTF8              string `json:"utf8"`
	ExternalReference string `json:"external_reference,omitempty"`
}

func buildReactionClassificationRows(resp jsonAPIResponse) []reactionClassificationRow {
	rows := make([]reactionClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := reactionClassificationRow{
			ID:                resource.ID,
			Label:             stringAttr(resource.Attributes, "label"),
			UTF8:              stringAttr(resource.Attributes, "utf8"),
			ExternalReference: stringAttr(resource.Attributes, "external-reference"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderReactionClassificationsTable(cmd *cobra.Command, rows []reactionClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No reaction classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLABEL\tUTF8\tEXTERNAL REFERENCE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Label, 20),
			row.UTF8,
			truncateString(row.ExternalReference, 30),
		)
	}
	return writer.Flush()
}
