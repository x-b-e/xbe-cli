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

type doProfferLikesCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Proffer string
	User    string
}

func newDoProfferLikesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a proffer like",
		Long: `Create a new proffer like.

Required flags:
  --proffer   Proffer ID (required)
  --user      User ID (required; must match the current user)`,
		Example: `  # Like a proffer
  xbe do proffer-likes create --proffer 123 --user 456

  # Get JSON output
  xbe do proffer-likes create --proffer 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProfferLikesCreate,
	}
	initDoProfferLikesCreateFlags(cmd)
	return cmd
}

func init() {
	doProfferLikesCmd.AddCommand(newDoProfferLikesCreateCmd())
}

func initDoProfferLikesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("proffer", "", "Proffer ID (required)")
	cmd.Flags().String("user", "", "User ID (required, must match current user)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProfferLikesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProfferLikesCreateOptions(cmd)
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

	if opts.Proffer == "" {
		err := fmt.Errorf("--proffer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"proffer": map[string]any{
			"data": map[string]any{
				"type": "proffers",
				"id":   opts.Proffer,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "proffer-likes",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/proffer-likes", jsonBody)
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

	row := buildProfferLikeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created proffer like %s\n", row.ID)
	return nil
}

func parseDoProfferLikesCreateOptions(cmd *cobra.Command) (doProfferLikesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	proffer, _ := cmd.Flags().GetString("proffer")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProfferLikesCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Proffer: proffer,
		User:    user,
	}, nil
}
