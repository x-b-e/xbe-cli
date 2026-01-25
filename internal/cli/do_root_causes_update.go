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

type doRootCausesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Title       string
	Description string
	IsTriaged   bool
}

func newDoRootCausesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing root cause",
		Long: `Update an existing root cause.

Provide the root cause ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --title        Title
  --description  Description
  --is-triaged   Mark as triaged (true/false)`,
		Example: `  # Update title
  xbe do root-causes update 123 --title "Updated root cause"

  # Update description
  xbe do root-causes update 123 --description "Updated description"

  # Update triage status
  xbe do root-causes update 123 --is-triaged false

  # JSON output
  xbe do root-causes update 123 --title "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRootCausesUpdate,
	}
	initDoRootCausesUpdateFlags(cmd)
	return cmd
}

func init() {
	doRootCausesCmd.AddCommand(newDoRootCausesUpdateCmd())
}

func initDoRootCausesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Root cause title")
	cmd.Flags().String("description", "", "Root cause description")
	cmd.Flags().Bool("is-triaged", false, "Mark as triaged")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRootCausesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRootCausesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-triaged") {
		attributes["is-triaged"] = opts.IsTriaged
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --title, --description, --is-triaged")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "root-causes",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/root-causes/"+opts.ID, jsonBody)
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

	row := buildRootCauseRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Title != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated root cause %s (%s)\n", row.ID, row.Title)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Updated root cause %s\n", row.ID)
	return nil
}

func parseDoRootCausesUpdateOptions(cmd *cobra.Command, args []string) (doRootCausesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	isTriaged, _ := cmd.Flags().GetBool("is-triaged")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRootCausesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Title:       title,
		Description: description,
		IsTriaged:   isTriaged,
	}, nil
}
