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

type doTagCategoriesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	Slug        string
	Description string
	CanApplyTo  []string
}

func newDoTagCategoriesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tag category",
		Long: `Update an existing tag category.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The tag category ID (required)

Flags:
  --name          Update the name
  --slug          Update the slug
  --description   Update the description
  --can-apply-to  Update entity types this can apply to`,
		Example: `  # Update just the name
  xbe do tag-categories update 123 --name "Updated Name"

  # Update description
  xbe do tag-categories update 123 --description "New description"

  # Update can-apply-to types
  xbe do tag-categories update 123 --can-apply-to PredictionSubject,Comment,Post

  # Get JSON output
  xbe do tag-categories update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTagCategoriesUpdate,
	}
	initDoTagCategoriesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTagCategoriesCmd.AddCommand(newDoTagCategoriesUpdateCmd())
}

func initDoTagCategoriesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("slug", "", "New slug")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().StringSlice("can-apply-to", nil, "New entity types this can apply to")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTagCategoriesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTagCategoriesUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("tag category id is required")
	}

	// Check if at least one field is being updated
	hasUpdate := opts.Name != "" || opts.Slug != "" || opts.Description != "" ||
		cmd.Flags().Changed("can-apply-to")

	if !hasUpdate {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Slug != "" {
		attributes["slug"] = opts.Slug
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("can-apply-to") {
		attributes["can-apply-to"] = opts.CanApplyTo
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "tag-categories",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tag-categories/"+id, jsonBody)
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

	row := buildTagCategoryRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tag category %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTagCategoriesUpdateOptions(cmd *cobra.Command) (doTagCategoriesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	slug, _ := cmd.Flags().GetString("slug")
	description, _ := cmd.Flags().GetString("description")
	canApplyTo, _ := cmd.Flags().GetStringSlice("can-apply-to")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTagCategoriesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		Slug:        slug,
		Description: description,
		CanApplyTo:  canApplyTo,
	}, nil
}
