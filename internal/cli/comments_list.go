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

type commentsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	CommentableType          string
	CommentableID            string
	CreatedBy                string
	DriverDayDriver          string
	JobProductionPlanProject string
}

type commentRow struct {
	ID              string `json:"id"`
	Body            string `json:"body,omitempty"`
	IsAdminOnly     bool   `json:"is_admin_only"`
	DoNotNotify     bool   `json:"do_not_notify"`
	IncludeInRecap  bool   `json:"include_in_recap"`
	IsCreatedByBot  bool   `json:"is_created_by_bot"`
	CommentableType string `json:"commentable_type,omitempty"`
	CommentableID   string `json:"commentable_id,omitempty"`
	CreatedByID     string `json:"created_by_id,omitempty"`
}

func newCommentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments",
		Long: `List comments on various resources.

Output Columns:
  ID               Comment identifier
  BODY             Comment text (truncated)
  ADMIN ONLY       Whether comment is admin-only
  COMMENTABLE      Type and ID of commented resource
  CREATED BY       User who created the comment

Filters:
  --commentable-type          Filter by commentable type (e.g., projects, truckers)
  --commentable-id            Filter by commentable ID
  --created-by                Filter by created-by user ID
  --driver-day-driver         Filter by driver day driver ID
  --job-production-plan-project  Filter by job production plan project ID`,
		Example: `  # List all comments
  xbe view comments list

  # Filter by commentable
  xbe view comments list --commentable-type projects --commentable-id 123

  # Filter by creator
  xbe view comments list --created-by 456

  # Output as JSON
  xbe view comments list --json`,
		RunE: runCommentsList,
	}
	initCommentsListFlags(cmd)
	return cmd
}

func init() {
	commentsCmd.AddCommand(newCommentsListCmd())
}

func initCommentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("commentable-type", "", "Filter by commentable type")
	cmd.Flags().String("commentable-id", "", "Filter by commentable ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("driver-day-driver", "", "Filter by driver day driver ID")
	cmd.Flags().String("job-production-plan-project", "", "Filter by job production plan project ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommentsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Handle polymorphic commentable filter
	if opts.CommentableType != "" && opts.CommentableID != "" {
		query.Set("filter[commentable]", opts.CommentableType+"|"+opts.CommentableID)
	} else if opts.CommentableType != "" {
		query.Set("filter[commentable_type]", opts.CommentableType)
	}
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[driver_day_driver]", opts.DriverDayDriver)
	setFilterIfPresent(query, "filter[job_production_plan_project]", opts.JobProductionPlanProject)

	body, _, err := client.Get(cmd.Context(), "/v1/comments", query)
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

	rows := buildCommentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommentsTable(cmd, rows)
}

func parseCommentsListOptions(cmd *cobra.Command) (commentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	commentableType, _ := cmd.Flags().GetString("commentable-type")
	commentableID, _ := cmd.Flags().GetString("commentable-id")
	createdBy, _ := cmd.Flags().GetString("created-by")
	driverDayDriver, _ := cmd.Flags().GetString("driver-day-driver")
	jobProductionPlanProject, _ := cmd.Flags().GetString("job-production-plan-project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commentsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		CommentableType:          commentableType,
		CommentableID:            commentableID,
		CreatedBy:                createdBy,
		DriverDayDriver:          driverDayDriver,
		JobProductionPlanProject: jobProductionPlanProject,
	}, nil
}

func buildCommentRows(resp jsonAPIResponse) []commentRow {
	rows := make([]commentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := commentRow{
			ID:             resource.ID,
			Body:           stringAttr(resource.Attributes, "body"),
			IsAdminOnly:    boolAttr(resource.Attributes, "is-admin-only"),
			DoNotNotify:    boolAttr(resource.Attributes, "do-not-notify"),
			IncludeInRecap: boolAttr(resource.Attributes, "include-in-recap"),
			IsCreatedByBot: boolAttr(resource.Attributes, "is-created-by-bot"),
		}

		if rel, ok := resource.Relationships["commentable"]; ok && rel.Data != nil {
			row.CommentableType = rel.Data.Type
			row.CommentableID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCommentsTable(cmd *cobra.Command, rows []commentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No comments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBODY\tADMIN ONLY\tCOMMENTABLE\tCREATED BY")
	for _, row := range rows {
		adminOnly := "no"
		if row.IsAdminOnly {
			adminOnly = "yes"
		}
		commentable := ""
		if row.CommentableType != "" && row.CommentableID != "" {
			commentable = row.CommentableType + "/" + row.CommentableID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Body, 40),
			adminOnly,
			commentable,
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
