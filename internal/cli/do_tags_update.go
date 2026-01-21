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

type doTagsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Name    string
}

func newDoTagsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing tag",
		Long: `Update an existing tag.

Provide the tag ID as an argument, then use flags to specify which
fields to update.

Updatable fields:
  --name  The tag name

Note: The tag category cannot be changed after creation.`,
		Example: `  # Update name
  xbe do tags update 123 --name "High Priority"

  # Get JSON output
  xbe do tags update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTagsUpdate,
	}
	initDoTagsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTagsCmd.AddCommand(newDoTagsUpdateCmd())
}

func initDoTagsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Tag name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTagsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTagsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --name")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tags",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/tags/"+opts.ID, jsonBody)
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

	row := buildTagRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tag %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTagsUpdateOptions(cmd *cobra.Command, args []string) (doTagsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTagsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Name:    name,
	}, nil
}
