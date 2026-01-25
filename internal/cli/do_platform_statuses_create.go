package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPlatformStatusesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Title       string
	Description string
	PublishedAt string
	StartAt     string
	EndAt       string
}

func newDoPlatformStatusesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new platform status",
		Long: `Create a new platform status update.

Required flags:
  --title         Status title
  --description   Status description

Optional flags:
  --published-at  Published timestamp (ISO 8601)
  --start-at      Start timestamp (ISO 8601)
  --end-at        End timestamp (ISO 8601)`,
		Example: `  # Create a platform status
  xbe do platform-statuses create --title "API Maintenance" --description "The API will be unavailable."

  # Create with scheduling details
  xbe do platform-statuses create --title "Planned Maintenance" --description "Scheduled downtime" --start-at 2024-05-01T01:00:00Z --end-at 2024-05-01T03:00:00Z`,
		RunE: runDoPlatformStatusesCreate,
	}
	initDoPlatformStatusesCreateFlags(cmd)
	return cmd
}

func init() {
	doPlatformStatusesCmd.AddCommand(newDoPlatformStatusesCreateCmd())
}

func initDoPlatformStatusesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Status title (required)")
	cmd.Flags().String("description", "", "Status description (required)")
	cmd.Flags().String("published-at", "", "Published timestamp (ISO 8601)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("description")
}

func runDoPlatformStatusesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPlatformStatusesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	attributes := map[string]any{
		"title":       opts.Title,
		"description": opts.Description,
	}

	if opts.PublishedAt != "" {
		attributes["published-at"] = opts.PublishedAt
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}

	data := map[string]any{
		"type":       "platform-statuses",
		"attributes": attributes,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/platform-statuses", jsonBody)
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

	row := platformStatusRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created platform status %s (%s)\n", row.ID, row.Title)
	return nil
}

func parseDoPlatformStatusesCreateOptions(cmd *cobra.Command) (doPlatformStatusesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	publishedAt, _ := cmd.Flags().GetString("published-at")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPlatformStatusesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Title:       title,
		Description: description,
		PublishedAt: publishedAt,
		StartAt:     startAt,
		EndAt:       endAt,
	}, nil
}
