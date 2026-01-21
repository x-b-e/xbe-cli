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

type doTagCategoriesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	Slug        string
	Description string
	CanApplyTo  []string
}

func newDoTagCategoriesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tag category",
		Long: `Create a new tag category.

Required flags:
  --name          The tag category name (required)
  --slug          URL-friendly identifier (required)
  --can-apply-to  Entity types tags in this category can apply to (required, comma-separated)

Optional flags:
  --description   Description of the tag category`,
		Example: `  # Create a tag category for predictions
  xbe do tag-categories create --name "Market Area" --slug "market-area" --can-apply-to PredictionSubject

  # Create with multiple apply-to types
  xbe do tag-categories create --name "Sentiment" --slug "sentiment" --can-apply-to PredictionSubject,Comment

  # Create with description
  xbe do tag-categories create --name "Topics" --slug "topics" --can-apply-to Post --description "Post topic tags"

  # Get JSON output
  xbe do tag-categories create --name "Test" --slug "test" --can-apply-to Comment --json`,
		Args: cobra.NoArgs,
		RunE: runDoTagCategoriesCreate,
	}
	initDoTagCategoriesCreateFlags(cmd)
	return cmd
}

func init() {
	doTagCategoriesCmd.AddCommand(newDoTagCategoriesCreateCmd())
}

func initDoTagCategoriesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Tag category name (required)")
	cmd.Flags().String("slug", "", "URL-friendly identifier (required)")
	cmd.Flags().String("description", "", "Description of the tag category")
	cmd.Flags().StringSlice("can-apply-to", nil, "Entity types this can apply to (required, comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTagCategoriesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTagCategoriesCreateOptions(cmd)
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require slug
	if opts.Slug == "" {
		err := fmt.Errorf("--slug is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require can-apply-to
	if len(opts.CanApplyTo) == 0 {
		err := fmt.Errorf("--can-apply-to is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name":         opts.Name,
		"slug":         opts.Slug,
		"can-apply-to": opts.CanApplyTo,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	requestBody := map[string]any{
		"data": map[string]any{
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

	body, _, err := client.Post(cmd.Context(), "/v1/tag-categories", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created tag category %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTagCategoriesCreateOptions(cmd *cobra.Command) (doTagCategoriesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	slug, _ := cmd.Flags().GetString("slug")
	description, _ := cmd.Flags().GetString("description")
	canApplyTo, _ := cmd.Flags().GetStringSlice("can-apply-to")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTagCategoriesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		Slug:        slug,
		Description: description,
		CanApplyTo:  canApplyTo,
	}, nil
}

type tagCategoryRow struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description,omitempty"`
	CanApplyTo  []string `json:"can_apply_to,omitempty"`
}

func buildTagCategoryRow(resp jsonAPISingleResponse) tagCategoryRow {
	attrs := resp.Data.Attributes

	return tagCategoryRow{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		Slug:        stringAttr(attrs, "slug"),
		Description: stringAttr(attrs, "description"),
		CanApplyTo:  stringSliceAttr(attrs, "can-apply-to"),
	}
}
