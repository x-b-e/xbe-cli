package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type taggingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type taggingDetails struct {
	ID           string `json:"id"`
	TagID        string `json:"tag_id,omitempty"`
	TagName      string `json:"tag_name,omitempty"`
	TaggableType string `json:"taggable_type,omitempty"`
	TaggableID   string `json:"taggable_id,omitempty"`
}

func newTaggingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tagging details",
		Long: `Show the full details of a tagging.

Output Fields:
  ID        Tagging identifier
  Tag       Tag linked to the tagging
  Taggable  Taggable type and ID

Arguments:
  <id>    Tagging ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tagging
  xbe view taggings show 123

  # JSON output
  xbe view taggings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTaggingsShow,
	}
	initTaggingsShowFlags(cmd)
	return cmd
}

func init() {
	taggingsCmd.AddCommand(newTaggingsShowCmd())
}

func initTaggingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTaggingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTaggingsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("tagging id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[taggings]", "tag,taggable")
	query.Set("include", "tag")
	query.Set("fields[tags]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/taggings/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTaggingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTaggingDetails(cmd, details)
}

func parseTaggingsShowOptions(cmd *cobra.Command) (taggingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return taggingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTaggingDetails(resp jsonAPISingleResponse) taggingDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := taggingDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["tag"]; ok && rel.Data != nil {
		details.TagID = rel.Data.ID
		if tag, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TagName = strings.TrimSpace(stringAttr(tag.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["taggable"]; ok && rel.Data != nil {
		details.TaggableType = rel.Data.Type
		details.TaggableID = rel.Data.ID
	}

	return details
}

func renderTaggingDetails(cmd *cobra.Command, details taggingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	writeLabelWithID(out, "Tag", details.TagName, details.TagID)

	taggable := ""
	if details.TaggableType != "" && details.TaggableID != "" {
		taggable = details.TaggableType + "/" + details.TaggableID
	} else if details.TaggableType != "" {
		taggable = details.TaggableType
	} else if details.TaggableID != "" {
		taggable = details.TaggableID
	}
	if taggable != "" {
		fmt.Fprintf(out, "Taggable: %s\n", taggable)
	}

	return nil
}
