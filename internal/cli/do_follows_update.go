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

type doFollowsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Follower    string
	Creator     string
	CreatorType string
	CreatorID   string
}

func newDoFollowsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a follow",
		Long: `Update a follow relationship.

Optional flags:
  --follower       Follower user ID
  --creator-type   Creator type (resource type, e.g., projects)
  --creator-id     Creator ID
  --creator        Creator type and ID (e.g., projects|123)`,
		Example: `  # Update creator
  xbe do follows update 123 --creator-type projects --creator-id 456

  # Update follower
  xbe do follows update 123 --follower 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoFollowsUpdate,
	}
	initDoFollowsUpdateFlags(cmd)
	return cmd
}

func init() {
	doFollowsCmd.AddCommand(newDoFollowsUpdateCmd())
}

func initDoFollowsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("follower", "", "Follower user ID")
	cmd.Flags().String("creator", "", "Creator (resource type|ID, e.g., projects|123)")
	cmd.Flags().String("creator-type", "", "Creator type (resource type, e.g., projects)")
	cmd.Flags().String("creator-id", "", "Creator ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoFollowsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoFollowsUpdateOptions(cmd, args)
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

	if opts.Creator != "" && (opts.CreatorType != "" || opts.CreatorID != "") {
		err := fmt.Errorf("--creator cannot be combined with --creator-type or --creator-id")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{}

	if strings.TrimSpace(opts.Follower) != "" {
		relationships["follower"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Follower,
			},
		}
	}

	creatorType := strings.TrimSpace(opts.CreatorType)
	creatorID := strings.TrimSpace(opts.CreatorID)
	if opts.Creator != "" {
		parts := strings.SplitN(opts.Creator, "|", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			err := fmt.Errorf("--creator must be in the format type|id")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		creatorType = strings.TrimSpace(parts[0])
		creatorID = strings.TrimSpace(parts[1])
	}

	if creatorType != "" || creatorID != "" {
		if creatorType == "" || creatorID == "" {
			err := fmt.Errorf("--creator-type and --creator-id must be used together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["creator"] = map[string]any{
			"data": map[string]any{
				"type": creatorType,
				"id":   creatorID,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "follows",
			"id":            opts.ID,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/follows/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated follow %s\n", row.ID)
	return nil
}

func parseDoFollowsUpdateOptions(cmd *cobra.Command, args []string) (doFollowsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	follower, _ := cmd.Flags().GetString("follower")
	creator, _ := cmd.Flags().GetString("creator")
	creatorType, _ := cmd.Flags().GetString("creator-type")
	creatorID, _ := cmd.Flags().GetString("creator-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFollowsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Follower:    follower,
		Creator:     creator,
		CreatorType: creatorType,
		CreatorID:   creatorID,
	}, nil
}
