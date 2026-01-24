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

type doPublicPraiseCultureValuesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	PublicPraise string
	CultureValue string
}

func newDoPublicPraiseCultureValuesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a public praise culture value",
		Long: `Create a new public praise culture value link.

Required flags:
  --public-praise  Public praise ID (required)
  --culture-value  Culture value ID (required)`,
		Example: `  # Link a public praise to a culture value
  xbe do public-praise-culture-values create --public-praise 123 --culture-value 456

  # Get JSON output
  xbe do public-praise-culture-values create --public-praise 123 --culture-value 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPublicPraiseCultureValuesCreate,
	}
	initDoPublicPraiseCultureValuesCreateFlags(cmd)
	return cmd
}

func init() {
	doPublicPraiseCultureValuesCmd.AddCommand(newDoPublicPraiseCultureValuesCreateCmd())
}

func initDoPublicPraiseCultureValuesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("public-praise", "", "Public praise ID (required)")
	cmd.Flags().String("culture-value", "", "Culture value ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPublicPraiseCultureValuesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPublicPraiseCultureValuesCreateOptions(cmd)
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

	if opts.PublicPraise == "" {
		err := fmt.Errorf("--public-praise is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.CultureValue == "" {
		err := fmt.Errorf("--culture-value is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"public-praise": map[string]any{
			"data": map[string]any{
				"type": "public-praises",
				"id":   opts.PublicPraise,
			},
		},
		"culture-value": map[string]any{
			"data": map[string]any{
				"type": "culture-values",
				"id":   opts.CultureValue,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "public-praise-culture-values",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/public-praise-culture-values", jsonBody)
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

	row := buildPublicPraiseCultureValueRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created public praise culture value %s\n", row.ID)
	return nil
}

func parseDoPublicPraiseCultureValuesCreateOptions(cmd *cobra.Command) (doPublicPraiseCultureValuesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	publicPraise, _ := cmd.Flags().GetString("public-praise")
	cultureValue, _ := cmd.Flags().GetString("culture-value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPublicPraiseCultureValuesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		PublicPraise: publicPraise,
		CultureValue: cultureValue,
	}, nil
}
