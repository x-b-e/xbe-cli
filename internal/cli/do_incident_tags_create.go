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

type doIncidentTagsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Slug        string
	Name        string
	Description string
	Kinds       string
}

func newDoIncidentTagsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new incident tag",
		Long: `Create a new incident tag.

Required flags:
  --slug         Unique slug identifier
  --name         Tag name

Optional flags:
  --description  Tag description
  --kinds        Kinds (comma-separated)`,
		Example: `  # Create an incident tag
  xbe do incident-tags create --slug "safety-violation" --name "Safety Violation"

  # Create with description
  xbe do incident-tags create --slug "delay" --name "Delay" --description "Production delay incident"`,
		RunE: runDoIncidentTagsCreate,
	}
	initDoIncidentTagsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentTagsCmd.AddCommand(newDoIncidentTagsCreateCmd())
}

func initDoIncidentTagsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("slug", "", "Unique slug identifier (required)")
	cmd.Flags().String("name", "", "Tag name (required)")
	cmd.Flags().String("description", "", "Tag description")
	cmd.Flags().String("kinds", "", "Kinds (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("slug")
	cmd.MarkFlagRequired("name")
}

func runDoIncidentTagsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentTagsCreateOptions(cmd)
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
		"slug": opts.Slug,
		"name": opts.Name,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Kinds != "" {
		attributes["kinds"] = strings.Split(opts.Kinds, ",")
	}

	data := map[string]any{
		"type":       "incident-tags",
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

	body, _, err := client.Post(cmd.Context(), "/v1/incident-tags", jsonBody)
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
			"id":   resp.Data.ID,
			"slug": stringAttr(resp.Data.Attributes, "slug"),
			"name": stringAttr(resp.Data.Attributes, "name"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident tag %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "name"))
	return nil
}

func parseDoIncidentTagsCreateOptions(cmd *cobra.Command) (doIncidentTagsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	slug, _ := cmd.Flags().GetString("slug")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	kinds, _ := cmd.Flags().GetString("kinds")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentTagsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Slug:        slug,
		Name:        name,
		Description: description,
		Kinds:       kinds,
	}, nil
}
