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

type openAiRealtimeSessionsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	ClientFeature            string
	User                     string
	ClientSecretExpiresAtMin string
	ClientSecretExpiresAtMax string
	IsClientSecretExpiresAt  string
	CreatedAtMin             string
	CreatedAtMax             string
	IsCreatedAt              string
	UpdatedAtMin             string
	UpdatedAtMax             string
	IsUpdatedAt              string
}

type openAiRealtimeSessionRow struct {
	ID                    string `json:"id"`
	Model                 string `json:"model,omitempty"`
	ClientFeature         string `json:"client_feature,omitempty"`
	UserID                string `json:"user_id,omitempty"`
	ClientSecretExpiresAt string `json:"client_secret_expires_at,omitempty"`
	Error                 string `json:"error,omitempty"`
}

func newOpenAiRealtimeSessionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OpenAI realtime sessions",
		Long: `List OpenAI realtime sessions.

Requires admin privileges.

Output Columns:
  ID         Session identifier
  MODEL      OpenAI model
  FEATURE    Client feature enum
  USER       User ID (if set)
  EXPIRES AT Client secret expiration timestamp
  ERROR      Error message (if any)

Filters:
  --client-feature                 Filter by client feature enum
  --user                           Filter by user ID
  --client-secret-expires-at-min   Filter by minimum secret expiration time (ISO 8601)
  --client-secret-expires-at-max   Filter by maximum secret expiration time (ISO 8601)
  --is-client-secret-expires-at    Filter by whether secret expiration is set (true/false)
  --created-at-min                 Filter by created-at on/after (ISO 8601)
  --created-at-max                 Filter by created-at on/before (ISO 8601)
  --is-created-at                  Filter by has created-at (true/false)
  --updated-at-min                 Filter by updated-at on/after (ISO 8601)
  --updated-at-max                 Filter by updated-at on/before (ISO 8601)
  --is-updated-at                  Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List sessions
  xbe view open-ai-realtime-sessions list

  # Filter by client feature
  xbe view open-ai-realtime-sessions list --client-feature giant_anchor_prediction_creation

  # Filter by user
  xbe view open-ai-realtime-sessions list --user 123

  # Filter by expiration window
  xbe view open-ai-realtime-sessions list --client-secret-expires-at-min 2025-01-01T00:00:00Z --client-secret-expires-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view open-ai-realtime-sessions list --json`,
		Args: cobra.NoArgs,
		RunE: runOpenAiRealtimeSessionsList,
	}
	initOpenAiRealtimeSessionsListFlags(cmd)
	return cmd
}

func init() {
	openAiRealtimeSessionsCmd.AddCommand(newOpenAiRealtimeSessionsListCmd())
}

func initOpenAiRealtimeSessionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("client-feature", "", "Filter by client feature enum")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("client-secret-expires-at-min", "", "Filter by minimum secret expiration time (ISO 8601)")
	cmd.Flags().String("client-secret-expires-at-max", "", "Filter by maximum secret expiration time (ISO 8601)")
	cmd.Flags().String("is-client-secret-expires-at", "", "Filter by whether secret expiration is set (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenAiRealtimeSessionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOpenAiRealtimeSessionsListOptions(cmd)
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
	query.Set("fields[open-ai-realtime-sessions]", "model,client-feature,client-secret-expires-at,error,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	setFilterIfPresent(query, "filter[client_feature]", opts.ClientFeature)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[client_secret_expires_at_min]", opts.ClientSecretExpiresAtMin)
	setFilterIfPresent(query, "filter[client_secret_expires_at_max]", opts.ClientSecretExpiresAtMax)
	setFilterIfPresent(query, "filter[is_client_secret_expires_at]", opts.IsClientSecretExpiresAt)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/open-ai-realtime-sessions", query)
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

	rows := buildOpenAiRealtimeSessionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOpenAiRealtimeSessionsTable(cmd, rows)
}

func parseOpenAiRealtimeSessionsListOptions(cmd *cobra.Command) (openAiRealtimeSessionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	clientFeature, _ := cmd.Flags().GetString("client-feature")
	user, _ := cmd.Flags().GetString("user")
	clientSecretExpiresAtMin, _ := cmd.Flags().GetString("client-secret-expires-at-min")
	clientSecretExpiresAtMax, _ := cmd.Flags().GetString("client-secret-expires-at-max")
	isClientSecretExpiresAt, _ := cmd.Flags().GetString("is-client-secret-expires-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openAiRealtimeSessionsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		ClientFeature:            clientFeature,
		User:                     user,
		ClientSecretExpiresAtMin: clientSecretExpiresAtMin,
		ClientSecretExpiresAtMax: clientSecretExpiresAtMax,
		IsClientSecretExpiresAt:  isClientSecretExpiresAt,
		CreatedAtMin:             createdAtMin,
		CreatedAtMax:             createdAtMax,
		IsCreatedAt:              isCreatedAt,
		UpdatedAtMin:             updatedAtMin,
		UpdatedAtMax:             updatedAtMax,
		IsUpdatedAt:              isUpdatedAt,
	}, nil
}

func buildOpenAiRealtimeSessionRows(resp jsonAPIResponse) []openAiRealtimeSessionRow {
	rows := make([]openAiRealtimeSessionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := openAiRealtimeSessionRow{
			ID:                    resource.ID,
			Model:                 strings.TrimSpace(stringAttr(resource.Attributes, "model")),
			ClientFeature:         strings.TrimSpace(stringAttr(resource.Attributes, "client-feature")),
			ClientSecretExpiresAt: formatDateTime(stringAttr(resource.Attributes, "client-secret-expires-at")),
			Error:                 strings.TrimSpace(stringAttr(resource.Attributes, "error")),
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderOpenAiRealtimeSessionsTable(cmd *cobra.Command, rows []openAiRealtimeSessionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No OpenAI realtime sessions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMODEL\tFEATURE\tUSER\tEXPIRES AT\tERROR")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Model,
			row.ClientFeature,
			row.UserID,
			row.ClientSecretExpiresAt,
			row.Error,
		)
	}
	return writer.Flush()
}
