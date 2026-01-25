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

type doDeveloperCertifiedWeighersCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Developer string
	User      string
	Number    string
	IsActive  string
}

func newDoDeveloperCertifiedWeighersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a developer certified weigher",
		Long: `Create a developer certified weigher.

Required flags:
  --developer   Developer ID (required)
  --user        User ID (required)

Optional flags:
  --number      Certified weigher number
  --is-active   Active status (true/false)

Note: The user must have a material supplier membership with the developer's broker.`,
		Example: `  # Create a developer certified weigher
  xbe do developer-certified-weighers create --developer 123 --user 456 --number CW-001

  # Create and set inactive
  xbe do developer-certified-weighers create --developer 123 --user 456 --is-active false

  # Output as JSON
  xbe do developer-certified-weighers create --developer 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoDeveloperCertifiedWeighersCreate,
	}
	initDoDeveloperCertifiedWeighersCreateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperCertifiedWeighersCmd.AddCommand(newDoDeveloperCertifiedWeighersCreateCmd())
}

func initDoDeveloperCertifiedWeighersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("developer", "", "Developer ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("number", "", "Certified weigher number")
	cmd.Flags().String("is-active", "", "Active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperCertifiedWeighersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDeveloperCertifiedWeighersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Developer) == "" {
		err := fmt.Errorf("--developer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.User) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("number") {
		attributes["number"] = opts.Number
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive == "true"
	}

	relationships := map[string]any{
		"developer": map[string]any{
			"data": map[string]any{
				"type": "developers",
				"id":   opts.Developer,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	data := map[string]any{
		"type":          "developer-certified-weighers",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/developer-certified-weighers", jsonBody)
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

	row := developerCertifiedWeigherRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer certified weigher %s\n", row.ID)
	return nil
}

func parseDoDeveloperCertifiedWeighersCreateOptions(cmd *cobra.Command) (doDeveloperCertifiedWeighersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	developer, _ := cmd.Flags().GetString("developer")
	user, _ := cmd.Flags().GetString("user")
	number, _ := cmd.Flags().GetString("number")
	isActive, _ := cmd.Flags().GetString("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperCertifiedWeighersCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Developer: developer,
		User:      user,
		Number:    number,
		IsActive:  isActive,
	}, nil
}
