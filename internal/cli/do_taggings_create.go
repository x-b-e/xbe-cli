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

type doTaggingsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	TagID        string
	TaggableType string
	TaggableID   string
}

func newDoTaggingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tagging",
		Long: `Create a tagging that links a tag to a taggable resource.

Required flags:
  --tag            Tag ID (required)
  --taggable-type  Taggable resource type (required, JSON API type such as prediction-subjects)
  --taggable-id    Taggable resource ID (required)

Note: The tag's category must allow the specified taggable type. Taggings are immutable after creation.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a tagging for a prediction subject
  xbe do taggings create \
    --tag 123 \
    --taggable-type prediction-subjects \
    --taggable-id 456

  # JSON output
  xbe do taggings create --tag 123 --taggable-type prediction-subjects --taggable-id 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTaggingsCreate,
	}
	initDoTaggingsCreateFlags(cmd)
	return cmd
}

func init() {
	doTaggingsCmd.AddCommand(newDoTaggingsCreateCmd())
}

func initDoTaggingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tag", "", "Tag ID (required)")
	cmd.Flags().String("taggable-type", "", "Taggable resource type (required)")
	cmd.Flags().String("taggable-id", "", "Taggable resource ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTaggingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTaggingsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TagID) == "" {
		err := fmt.Errorf("--tag is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TaggableType) == "" {
		err := fmt.Errorf("--taggable-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TaggableID) == "" {
		err := fmt.Errorf("--taggable-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"tag": map[string]any{
			"data": map[string]any{
				"type": "tags",
				"id":   opts.TagID,
			},
		},
		"taggable": map[string]any{
			"data": map[string]any{
				"type": opts.TaggableType,
				"id":   opts.TaggableID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "taggings",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/taggings", jsonBody)
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

	row := buildTaggingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	taggable := ""
	if row.TaggableType != "" && row.TaggableID != "" {
		taggable = row.TaggableType + "/" + row.TaggableID
	}
	if taggable != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created tagging %s (tag %s, taggable %s)\n", row.ID, row.TagID, taggable)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created tagging %s\n", row.ID)
	return nil
}

func parseDoTaggingsCreateOptions(cmd *cobra.Command) (doTaggingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tagID, _ := cmd.Flags().GetString("tag")
	taggableType, _ := cmd.Flags().GetString("taggable-type")
	taggableID, _ := cmd.Flags().GetString("taggable-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTaggingsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		TagID:        tagID,
		TaggableType: taggableType,
		TaggableID:   taggableID,
	}, nil
}

func buildTaggingRowFromSingle(resp jsonAPISingleResponse) taggingRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	row := taggingRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["tag"]; ok && rel.Data != nil {
		row.TagID = rel.Data.ID
		if tag, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TagName = strings.TrimSpace(stringAttr(tag.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["taggable"]; ok && rel.Data != nil {
		row.TaggableType = rel.Data.Type
		row.TaggableID = rel.Data.ID
	}

	return row
}
