package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type platformStatusesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type platformStatusDetails struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	StartAt     string `json:"start_at,omitempty"`
	EndAt       string `json:"end_at,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
}

func newPlatformStatusesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show platform status details",
		Long: `Show the full details of a specific platform status.

Retrieves and displays the complete status update, including the
description and scheduling details.

Output Fields:
  ID           Platform status identifier
  Title        Status title
  Published    Published timestamp
  Start        Start timestamp
  End          End timestamp
  Created By   User who created the status (if available)
  Description  Full status description

Arguments:
  <id>          The platform status ID (required).`,
		Example: `  # View a platform status by ID
  xbe view platform-statuses show 123

  # Get platform status as JSON
  xbe view platform-statuses show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPlatformStatusesShow,
	}
	initPlatformStatusesShowFlags(cmd)
	return cmd
}

func init() {
	platformStatusesCmd.AddCommand(newPlatformStatusesShowCmd())
}

func initPlatformStatusesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPlatformStatusesShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePlatformStatusesShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("platform status id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[platform-statuses]", "title,description,published-at,start-at,end-at")
	query.Set("fields[users]", "name")
	query.Set("include", "created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/platform-statuses/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildPlatformStatusDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPlatformStatusDetails(cmd, details)
}

func parsePlatformStatusesShowOptions(cmd *cobra.Command) (platformStatusesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return platformStatusesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return platformStatusesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return platformStatusesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return platformStatusesShowOptions{}, err
	}

	return platformStatusesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPlatformStatusDetails(resp jsonAPISingleResponse) platformStatusDetails {
	attrs := resp.Data.Attributes
	details := platformStatusDetails{
		ID:          resp.Data.ID,
		Title:       strings.TrimSpace(stringAttr(attrs, "title")),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
		PublishedAt: formatDateTime(stringAttr(attrs, "published-at")),
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
	}

	createdByType := ""
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.CreatedByID != "" && createdByType != "" {
		if user, ok := included[resourceKey(createdByType, details.CreatedByID)]; ok {
			details.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return details
}

func renderPlatformStatusDetails(cmd *cobra.Command, details platformStatusDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Title != "" {
		fmt.Fprintf(out, "Title: %s\n", details.Title)
	}
	if details.PublishedAt != "" {
		fmt.Fprintf(out, "Published: %s\n", details.PublishedAt)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End: %s\n", details.EndAt)
	}
	if details.CreatedBy != "" {
		if details.CreatedByID != "" {
			fmt.Fprintf(out, "Created By: %s (%s)\n", details.CreatedBy, details.CreatedByID)
		} else {
			fmt.Fprintf(out, "Created By: %s\n", details.CreatedBy)
		}
	} else if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}

	if details.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Description)
	}

	return nil
}
