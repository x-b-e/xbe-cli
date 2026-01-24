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

type doPublicPraiseCultureValuesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	PublicPraise string
	CultureValue string
}

func newDoPublicPraiseCultureValuesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a public praise culture value",
		Long: `Update an existing public praise culture value.

Arguments:
  <id>    The public praise culture value ID (required)

Flags:
  --public-praise  Public praise ID
  --culture-value  Culture value ID`,
		Example: `  # Update relationships
  xbe do public-praise-culture-values update 123 --public-praise 456 --culture-value 789

  # Get JSON output
  xbe do public-praise-culture-values update 123 --public-praise 456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPublicPraiseCultureValuesUpdate,
	}
	initDoPublicPraiseCultureValuesUpdateFlags(cmd)
	return cmd
}

func init() {
	doPublicPraiseCultureValuesCmd.AddCommand(newDoPublicPraiseCultureValuesUpdateCmd())
}

func initDoPublicPraiseCultureValuesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("public-praise", "", "Public praise ID")
	cmd.Flags().String("culture-value", "", "Culture value ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPublicPraiseCultureValuesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPublicPraiseCultureValuesUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("public-praise") {
		if strings.TrimSpace(opts.PublicPraise) == "" {
			err := fmt.Errorf("--public-praise cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["public-praise"] = map[string]any{
			"data": map[string]any{
				"type": "public-praises",
				"id":   opts.PublicPraise,
			},
		}
	}

	if cmd.Flags().Changed("culture-value") {
		if strings.TrimSpace(opts.CultureValue) == "" {
			err := fmt.Errorf("--culture-value cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["culture-value"] = map[string]any{
			"data": map[string]any{
				"type": "culture-values",
				"id":   opts.CultureValue,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "public-praise-culture-values",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/public-praise-culture-values/"+opts.ID, jsonBody)
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

	details := buildPublicPraiseCultureValueDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated public praise culture value %s\n", details.ID)
	return nil
}

func parseDoPublicPraiseCultureValuesUpdateOptions(cmd *cobra.Command, args []string) (doPublicPraiseCultureValuesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	publicPraise, _ := cmd.Flags().GetString("public-praise")
	cultureValue, _ := cmd.Flags().GetString("culture-value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPublicPraiseCultureValuesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		PublicPraise: publicPraise,
		CultureValue: cultureValue,
	}, nil
}
