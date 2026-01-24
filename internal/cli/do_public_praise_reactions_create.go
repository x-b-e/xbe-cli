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

type doPublicPraiseReactionsCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	PublicPraiseID           string
	ReactionClassificationID string
	CreatedByID              string
}

func newDoPublicPraiseReactionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a public praise reaction",
		Long: `Create a public praise reaction.

Required flags:
  --public-praise             Public praise ID (required)
  --reaction-classification   Reaction classification ID (required)

Optional flags:
  --created-by                Created-by user ID (admin only)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a public praise reaction
  xbe do public-praise-reactions create --public-praise 123 --reaction-classification 45

  # Create on behalf of another user (admin only)
  xbe do public-praise-reactions create --public-praise 123 --reaction-classification 45 --created-by 789

  # JSON output
  xbe do public-praise-reactions create --public-praise 123 --reaction-classification 45 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPublicPraiseReactionsCreate,
	}
	initDoPublicPraiseReactionsCreateFlags(cmd)
	return cmd
}

func init() {
	doPublicPraiseReactionsCmd.AddCommand(newDoPublicPraiseReactionsCreateCmd())
}

func initDoPublicPraiseReactionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("public-praise", "", "Public praise ID (required)")
	cmd.Flags().String("reaction-classification", "", "Reaction classification ID (required)")
	cmd.Flags().String("created-by", "", "Created-by user ID (admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPublicPraiseReactionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPublicPraiseReactionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.PublicPraiseID) == "" {
		err := fmt.Errorf("--public-praise is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.ReactionClassificationID) == "" {
		err := fmt.Errorf("--reaction-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"public-praise": map[string]any{
			"data": map[string]any{
				"type": "public-praises",
				"id":   opts.PublicPraiseID,
			},
		},
		"reaction-classification": map[string]any{
			"data": map[string]any{
				"type": "reaction-classifications",
				"id":   opts.ReactionClassificationID,
			},
		},
	}

	if strings.TrimSpace(opts.CreatedByID) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedByID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "public-praise-reactions",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/public-praise-reactions", jsonBody)
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

	row := buildPublicPraiseReactionRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.PublicPraiseID != "" && row.ReactionClassificationID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created public praise reaction %s for public praise %s (reaction %s)\n", row.ID, row.PublicPraiseID, row.ReactionClassificationID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created public praise reaction %s\n", row.ID)
	return nil
}

func parseDoPublicPraiseReactionsCreateOptions(cmd *cobra.Command) (doPublicPraiseReactionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	publicPraiseID, _ := cmd.Flags().GetString("public-praise")
	reactionClassificationID, _ := cmd.Flags().GetString("reaction-classification")
	createdByID, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPublicPraiseReactionsCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		PublicPraiseID:           publicPraiseID,
		ReactionClassificationID: reactionClassificationID,
		CreatedByID:              createdByID,
	}, nil
}
