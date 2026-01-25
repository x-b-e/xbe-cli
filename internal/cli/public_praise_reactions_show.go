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

type publicPraiseReactionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type publicPraiseReactionDetails struct {
	ID                       string `json:"id"`
	PublicPraiseID           string `json:"public_praise_id,omitempty"`
	ReactionClassificationID string `json:"reaction_classification_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
}

func newPublicPraiseReactionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show public praise reaction details",
		Long: `Show full details of a public praise reaction.

Output Fields:
  ID                       Reaction identifier
  Public Praise            Public praise ID
  Reaction Classification  Reaction classification ID
  Created By               User ID who created the reaction
  Created At               Creation timestamp
  Updated At               Update timestamp

Arguments:
  <id>    Public praise reaction ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a public praise reaction
  xbe view public-praise-reactions show 123

  # JSON output
  xbe view public-praise-reactions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPublicPraiseReactionsShow,
	}
	initPublicPraiseReactionsShowFlags(cmd)
	return cmd
}

func init() {
	publicPraiseReactionsCmd.AddCommand(newPublicPraiseReactionsShowCmd())
}

func initPublicPraiseReactionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPublicPraiseReactionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePublicPraiseReactionsShowOptions(cmd)
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
		return fmt.Errorf("public praise reaction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[public-praise-reactions]", "public-praise,created-by,reaction-classification,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/public-praise-reactions/"+id, query)
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

	details := buildPublicPraiseReactionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPublicPraiseReactionDetails(cmd, details)
}

func parsePublicPraiseReactionsShowOptions(cmd *cobra.Command) (publicPraiseReactionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return publicPraiseReactionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPublicPraiseReactionDetails(resp jsonAPISingleResponse) publicPraiseReactionDetails {
	attrs := resp.Data.Attributes
	details := publicPraiseReactionDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["public-praise"]; ok && rel.Data != nil {
		details.PublicPraiseID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["reaction-classification"]; ok && rel.Data != nil {
		details.ReactionClassificationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderPublicPraiseReactionDetails(cmd *cobra.Command, details publicPraiseReactionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Public Praise: %s\n", formatOptional(details.PublicPraiseID))
	fmt.Fprintf(out, "Reaction Classification: %s\n", formatOptional(details.ReactionClassificationID))
	fmt.Fprintf(out, "Created By: %s\n", formatOptional(details.CreatedByID))
	fmt.Fprintf(out, "Created At: %s\n", formatOptional(details.CreatedAt))
	fmt.Fprintf(out, "Updated At: %s\n", formatOptional(details.UpdatedAt))

	return nil
}
