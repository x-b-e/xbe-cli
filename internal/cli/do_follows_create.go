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

type doFollowsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Follower    string
	Creator     string
	CreatorType string
	CreatorID   string
}

func newDoFollowsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a follow",
		Long: `Create a follow relationship.

Required flags:
  --follower       Follower user ID (required)
  --creator-type   Creator type (resource type, e.g., projects) (required)
  --creator-id     Creator ID (required)

Alternative:
  --creator        Creator type and ID (e.g., projects|123)`,
		Example: `  # Follow a project
  xbe do follows create --follower 123 --creator-type projects --creator-id 456

  # Follow using combined creator flag
  xbe do follows create --follower 123 --creator projects|456`,
		Args: cobra.NoArgs,
		RunE: runDoFollowsCreate,
	}
	initDoFollowsCreateFlags(cmd)
	return cmd
}

func init() {
	doFollowsCmd.AddCommand(newDoFollowsCreateCmd())
}

func initDoFollowsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("follower", "", "Follower user ID (required)")
	cmd.Flags().String("creator", "", "Creator (resource type|ID, e.g., projects|123)")
	cmd.Flags().String("creator-type", "", "Creator type (resource type, e.g., projects)")
	cmd.Flags().String("creator-id", "", "Creator ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoFollowsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoFollowsCreateOptions(cmd)
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

	if opts.Follower == "" {
		err := fmt.Errorf("--follower is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	creatorType := strings.TrimSpace(opts.CreatorType)
	creatorID := strings.TrimSpace(opts.CreatorID)
	if opts.Creator != "" {
		if creatorType != "" || creatorID != "" {
			err := fmt.Errorf("--creator cannot be combined with --creator-type or --creator-id")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		parts := strings.SplitN(opts.Creator, "|", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			err := fmt.Errorf("--creator must be in the format type|id")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		creatorType = strings.TrimSpace(parts[0])
		creatorID = strings.TrimSpace(parts[1])
	}

	if creatorType == "" || creatorID == "" {
		err := fmt.Errorf("--creator-type and --creator-id are required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"follower": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Follower,
			},
		},
		"creator": map[string]any{
			"data": map[string]any{
				"type": creatorType,
				"id":   creatorID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "follows",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/follows", jsonBody)
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

	row := followRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created follow %s\n", row.ID)
	return nil
}

func parseDoFollowsCreateOptions(cmd *cobra.Command) (doFollowsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	follower, _ := cmd.Flags().GetString("follower")
	creator, _ := cmd.Flags().GetString("creator")
	creatorType, _ := cmd.Flags().GetString("creator-type")
	creatorID, _ := cmd.Flags().GetString("creator-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFollowsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Follower:    follower,
		Creator:     creator,
		CreatorType: creatorType,
		CreatorID:   creatorID,
	}, nil
}
