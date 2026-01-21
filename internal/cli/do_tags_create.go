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

type doTagsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	TagCategory string
}

func newDoTagsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tag",
		Long: `Create a new tag.

Required flags:
  --name          The tag name (required)
  --tag-category  The tag category ID (required)

Note: The tag category cannot be changed after creation.`,
		Example: `  # Create a tag
  xbe do tags create --name "Urgent" --tag-category 123

  # Get JSON output
  xbe do tags create --name "Important" --tag-category 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTagsCreate,
	}
	initDoTagsCreateFlags(cmd)
	return cmd
}

func init() {
	doTagsCmd.AddCommand(newDoTagsCreateCmd())
}

func initDoTagsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Tag name (required)")
	cmd.Flags().String("tag-category", "", "Tag category ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTagsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTagsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TagCategory == "" {
		err := fmt.Errorf("--tag-category is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tags",
			"attributes": attributes,
			"relationships": map[string]any{
				"tag-category": map[string]any{
					"data": map[string]any{
						"type": "tag-categories",
						"id":   opts.TagCategory,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tags", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created tag %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTagsCreateOptions(cmd *cobra.Command) (doTagsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	tagCategory, _ := cmd.Flags().GetString("tag-category")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTagsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		TagCategory: tagCategory,
	}, nil
}

func buildTagRowFromSingle(resp jsonAPISingleResponse) tagRow {
	attrs := resp.Data.Attributes

	row := tagRow{
		ID:   resp.Data.ID,
		Name: stringAttr(attrs, "name"),
	}

	if rel, ok := resp.Data.Relationships["tag-category"]; ok && rel.Data != nil {
		row.TagCategoryID = rel.Data.ID
	}

	return row
}
