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

type doProfitImprovementCategoriesCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Name    string
}

func newDoProfitImprovementCategoriesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new profit improvement category",
		Long: `Create a new profit improvement category.

Required flags:
  --name    Category name`,
		Example: `  # Create a profit improvement category
  xbe do profit-improvement-categories create --name "Fuel Savings"`,
		RunE: runDoProfitImprovementCategoriesCreate,
	}
	initDoProfitImprovementCategoriesCreateFlags(cmd)
	return cmd
}

func init() {
	doProfitImprovementCategoriesCmd.AddCommand(newDoProfitImprovementCategoriesCreateCmd())
}

func initDoProfitImprovementCategoriesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Category name (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
}

func runDoProfitImprovementCategoriesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProfitImprovementCategoriesCreateOptions(cmd)
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

	attributes := map[string]any{
		"name": opts.Name,
	}

	data := map[string]any{
		"type":       "profit-improvement-categories",
		"attributes": attributes,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/profit-improvement-categories", jsonBody)
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

	row := profitImprovementCategoryRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created profit improvement category %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProfitImprovementCategoriesCreateOptions(cmd *cobra.Command) (doProfitImprovementCategoriesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProfitImprovementCategoriesCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Name:    name,
	}, nil
}

func profitImprovementCategoryRowFromSingle(resp jsonAPISingleResponse) profitImprovementCategoryRow {
	return profitImprovementCategoryRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "name"),
	}
}
