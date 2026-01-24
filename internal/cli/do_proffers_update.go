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

type doProffersUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Title       string
	Description string
	Kind        string
}

func newDoProffersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a proffer",
		Long: `Update a proffer.

Optional flags:
  --title         Proffer title
  --description   Proffer description
  --kind          Proffer kind (hot_feed_post/make_it_so_action)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a proffer title
  xbe do proffers update 123 --title "Updated title"

  # Update description and kind
  xbe do proffers update 123 --description "More detail" --kind hot_feed_post`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProffersUpdate,
	}
	initDoProffersUpdateFlags(cmd)
	return cmd
}

func init() {
	doProffersCmd.AddCommand(newDoProffersUpdateCmd())
}

func initDoProffersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Proffer title")
	cmd.Flags().String("description", "", "Proffer description")
	cmd.Flags().String("kind", "", "Proffer kind (hot_feed_post/make_it_so_action)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProffersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProffersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "proffers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/proffers/"+opts.ID, jsonBody)
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

	row := profferRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated proffer %s (%s)\n", row.ID, row.Title)
	return nil
}

func parseDoProffersUpdateOptions(cmd *cobra.Command, args []string) (doProffersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProffersUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Title:       title,
		Description: description,
		Kind:        kind,
	}, nil
}

func profferRowFromSingle(resp jsonAPISingleResponse) profferRow {
	row := profferRow{
		ID:                  resp.Data.ID,
		Title:               strings.TrimSpace(stringAttr(resp.Data.Attributes, "title")),
		Kind:                stringAttr(resp.Data.Attributes, "kind"),
		CreatedByName:       stringAttr(resp.Data.Attributes, "created-by-name"),
		LikeCount:           intAttr(resp.Data.Attributes, "like-count"),
		ModerationStatus:    stringAttr(resp.Data.Attributes, "moderation-status"),
		Similarity:          stringAttr(resp.Data.Attributes, "similarity"),
		HasCurrentUserLiked: boolAttr(resp.Data.Attributes, "has-current-user-liked"),
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
