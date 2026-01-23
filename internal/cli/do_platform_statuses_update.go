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

type doPlatformStatusesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Title       string
	Description string
	PublishedAt string
	StartAt     string
	EndAt       string
}

func newDoPlatformStatusesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a platform status",
		Long: `Update an existing platform status update.

Optional flags:
  --title         Status title
  --description   Status description
  --published-at  Published timestamp (ISO 8601)
  --start-at      Start timestamp (ISO 8601)
  --end-at        End timestamp (ISO 8601)`,
		Example: `  # Update a platform status title
  xbe do platform-statuses update 123 --title "Updated Status"

  # Update scheduling details
  xbe do platform-statuses update 123 --start-at 2024-05-01T02:00:00Z --end-at 2024-05-01T04:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPlatformStatusesUpdate,
	}
	initDoPlatformStatusesUpdateFlags(cmd)
	return cmd
}

func init() {
	doPlatformStatusesCmd.AddCommand(newDoPlatformStatusesUpdateCmd())
}

func initDoPlatformStatusesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Status title")
	cmd.Flags().String("description", "", "Status description")
	cmd.Flags().String("published-at", "", "Published timestamp (ISO 8601)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPlatformStatusesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPlatformStatusesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}

	if cmd.Flags().Changed("title") {
		attributes["title"] = opts.Title
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("published-at") {
		attributes["published-at"] = opts.PublishedAt
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "platform-statuses",
		"id":         opts.ID,
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

	path := fmt.Sprintf("/v1/platform-statuses/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated platform status %s (%s)\n", row.ID, row.Title)
	return nil
}

func parseDoPlatformStatusesUpdateOptions(cmd *cobra.Command, args []string) (doPlatformStatusesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	publishedAt, _ := cmd.Flags().GetString("published-at")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPlatformStatusesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Title:       title,
		Description: description,
		PublishedAt: publishedAt,
		StartAt:     startAt,
		EndAt:       endAt,
	}, nil
}
