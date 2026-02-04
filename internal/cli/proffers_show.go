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

type proffersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type profferDetails struct {
	ID                  string   `json:"id"`
	Title               string   `json:"title,omitempty"`
	Description         string   `json:"description,omitempty"`
	Kind                string   `json:"kind,omitempty"`
	TitleGenerated      string   `json:"title_generated,omitempty"`
	CreatedByID         string   `json:"created_by_id,omitempty"`
	CreatedByName       string   `json:"created_by_name,omitempty"`
	LikeCount           int      `json:"like_count"`
	Similarity          string   `json:"similarity,omitempty"`
	HasCurrentUserLiked bool     `json:"has_current_user_liked"`
	ModerationStatus    string   `json:"moderation_status,omitempty"`
	ProfferLikeIDs      []string `json:"proffer_like_ids,omitempty"`
}

func newProffersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show proffer details",
		Long: `Show the full details of a specific proffer.

Output Fields:
  ID                    Proffer identifier
  Title                 User-provided title
  Description           Proffer description
  Kind                  Proffer kind
  Title Generated       AI-generated title
  Created By            Creator name
  Created By ID         Creator user ID
  Like Count            Number of likes
  Similarity            Similarity score (when using --similar-to-text)
  Has Current User Liked Whether the current user has liked the proffer
  Moderation Status     Moderation status
  Proffer Like IDs      Associated proffer-like IDs

Arguments:
  <id>    The proffer ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a proffer
  xbe view proffers show 123

  # JSON output
  xbe view proffers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProffersShow,
	}
	initProffersShowFlags(cmd)
	return cmd
}

func init() {
	proffersCmd.AddCommand(newProffersShowCmd())
}

func initProffersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProffersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProffersShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("proffer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[proffers]", "title,description,kind,title-generated,created-by-name,like-count,similarity,has-current-user-liked,moderation-status,created-by,proffer-likes")

	body, _, err := client.Get(cmd.Context(), "/v1/proffers/"+id, query)
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

	details := buildProfferDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProfferDetails(cmd, details)
}

func parseProffersShowOptions(cmd *cobra.Command) (proffersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return proffersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProfferDetails(resp jsonAPISingleResponse) profferDetails {
	attrs := resp.Data.Attributes
	details := profferDetails{
		ID:                  resp.Data.ID,
		Title:               strings.TrimSpace(stringAttr(attrs, "title")),
		Description:         strings.TrimSpace(stringAttr(attrs, "description")),
		Kind:                stringAttr(attrs, "kind"),
		TitleGenerated:      strings.TrimSpace(stringAttr(attrs, "title-generated")),
		CreatedByName:       stringAttr(attrs, "created-by-name"),
		LikeCount:           intAttr(attrs, "like-count"),
		Similarity:          stringAttr(attrs, "similarity"),
		HasCurrentUserLiked: boolAttr(attrs, "has-current-user-liked"),
		ModerationStatus:    stringAttr(attrs, "moderation-status"),
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["proffer-likes"]; ok {
		details.ProfferLikeIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProfferDetails(cmd *cobra.Command, details profferDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Title != "" {
		fmt.Fprintf(out, "Title: %s\n", details.Title)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.TitleGenerated != "" {
		fmt.Fprintf(out, "Title Generated: %s\n", details.TitleGenerated)
	}
	if details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByName)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}

	fmt.Fprintf(out, "Like Count: %d\n", details.LikeCount)
	fmt.Fprintf(out, "Has Current User Liked: %s\n", formatBool(details.HasCurrentUserLiked))

	if details.ModerationStatus != "" {
		fmt.Fprintf(out, "Moderation Status: %s\n", details.ModerationStatus)
	}
	if details.Similarity != "" {
		fmt.Fprintf(out, "Similarity: %s\n", details.Similarity)
	}
	if len(details.ProfferLikeIDs) > 0 {
		fmt.Fprintf(out, "Proffer Like IDs: %s\n", strings.Join(details.ProfferLikeIDs, ", "))
	}

	return nil
}
