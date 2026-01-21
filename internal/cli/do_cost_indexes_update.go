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

type doCostIndexesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Name        string
	Description string
	URL         string
	ExpiredAt   string
}

func newDoCostIndexesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing cost index",
		Long: `Update an existing cost index.

Provide the index ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name         The index name
  --description  Index description
  --url          Reference URL
  --expired-at   Expiration date (ISO 8601)`,
		Example: `  # Update name
  xbe do cost-indexes update 123 --name "Updated Name"

  # Update multiple fields
  xbe do cost-indexes update 123 --name "New Name" --description "New description"

  # Set expiration
  xbe do cost-indexes update 123 --expired-at "2025-12-31T00:00:00Z"

  # Get JSON output
  xbe do cost-indexes update 123 --name "New Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCostIndexesUpdate,
	}
	initDoCostIndexesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCostIndexesCmd.AddCommand(newDoCostIndexesUpdateCmd())
}

func initDoCostIndexesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Index name")
	cmd.Flags().String("description", "", "Index description")
	cmd.Flags().String("url", "", "Reference URL")
	cmd.Flags().String("expired-at", "", "Expiration date (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostIndexesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostIndexesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("url") {
		attributes["url"] = opts.URL
	}
	if cmd.Flags().Changed("expired-at") {
		attributes["expired-at"] = opts.ExpiredAt
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --description, --url, --expired-at")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "cost-indexes",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/cost-indexes/"+opts.ID, jsonBody)
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

	row := buildCostIndexRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated cost index %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCostIndexesUpdateOptions(cmd *cobra.Command, args []string) (doCostIndexesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	url, _ := cmd.Flags().GetString("url")
	expiredAt, _ := cmd.Flags().GetString("expired-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostIndexesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Name:        name,
		Description: description,
		URL:         url,
		ExpiredAt:   expiredAt,
	}, nil
}
