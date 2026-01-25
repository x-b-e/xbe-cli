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

type doProffersCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Title       string
	Description string
	Kind        string
}

func newDoProffersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new proffer",
		Long: `Create a new proffer (feature suggestion).

Required flags:
  --title         Proffer title

Optional flags:
  --description   Proffer description
  --kind          Proffer kind (hot_feed_post/make_it_so_action)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a proffer
  xbe do proffers create --title "Add CSV export"

  # Create with description and kind
  xbe do proffers create --title "Notification preferences" --description "Allow per-project settings" --kind make_it_so_action`,
		RunE: runDoProffersCreate,
	}
	initDoProffersCreateFlags(cmd)
	return cmd
}

func init() {
	doProffersCmd.AddCommand(newDoProffersCreateCmd())
}

func initDoProffersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Proffer title (required)")
	cmd.Flags().String("description", "", "Proffer description")
	cmd.Flags().String("kind", "", "Proffer kind (hot_feed_post/make_it_so_action)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("title")
}

func runDoProffersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProffersCreateOptions(cmd)
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
		"title": opts.Title,
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "proffers",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/proffers", jsonBody)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":    resp.Data.ID,
			"title": stringAttr(resp.Data.Attributes, "title"),
			"kind":  stringAttr(resp.Data.Attributes, "kind"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created proffer %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "title"))
	return nil
}

func parseDoProffersCreateOptions(cmd *cobra.Command) (doProffersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProffersCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Title:       title,
		Description: description,
		Kind:        kind,
	}, nil
}
