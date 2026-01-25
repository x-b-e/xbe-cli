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

type doProjectCategoriesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	Abbreviation string
	Broker       string
}

func newDoProjectCategoriesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project category",
		Long: `Create a new project category.

Required flags:
  --name         Category name
  --broker       Broker ID

Optional flags:
  --abbreviation Category abbreviation`,
		Example: `  # Create a project category
  xbe do project-categories create --name "Commercial" --broker 123

  # Create with abbreviation
  xbe do project-categories create --name "Residential" --abbreviation "RES" --broker 123`,
		RunE: runDoProjectCategoriesCreate,
	}
	initDoProjectCategoriesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectCategoriesCmd.AddCommand(newDoProjectCategoriesCreateCmd())
}

func initDoProjectCategoriesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Category name (required)")
	cmd.Flags().String("abbreviation", "", "Category abbreviation")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("broker")
}

func runDoProjectCategoriesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectCategoriesCreateOptions(cmd)
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

	if opts.Abbreviation != "" {
		attributes["abbreviation"] = opts.Abbreviation
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	data := map[string]any{
		"type":          "project-categories",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-categories", jsonBody)
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

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":   resp.Data.ID,
			"name": stringAttr(resp.Data.Attributes, "name"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project category %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "name"))
	return nil
}

func parseDoProjectCategoriesCreateOptions(cmd *cobra.Command) (doProjectCategoriesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCategoriesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		Abbreviation: abbreviation,
		Broker:       broker,
	}, nil
}
