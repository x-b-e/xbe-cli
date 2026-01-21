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

type doCostIndexesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	Description string
	URL         string
	ExpiredAt   string
	Broker      string
}

func newDoCostIndexesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cost index",
		Long: `Create a new cost index.

Required flags:
  --name  The index name (required)

Optional flags:
  --description  Index description
  --url          Reference URL for the index
  --expired-at   Expiration date (ISO 8601 format)
  --broker       Broker ID (omit for global index)`,
		Example: `  # Create a global cost index
  xbe do cost-indexes create --name "National CPI"

  # Create a broker-specific cost index
  xbe do cost-indexes create --name "Fuel Index" --broker 123

  # Create with all options
  xbe do cost-indexes create --name "Regional CPI" --description "Consumer Price Index" --url "https://example.com/cpi" --broker 123

  # Get JSON output
  xbe do cost-indexes create --name "Labor Index" --json`,
		Args: cobra.NoArgs,
		RunE: runDoCostIndexesCreate,
	}
	initDoCostIndexesCreateFlags(cmd)
	return cmd
}

func init() {
	doCostIndexesCmd.AddCommand(newDoCostIndexesCreateCmd())
}

func initDoCostIndexesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Index name (required)")
	cmd.Flags().String("description", "", "Index description")
	cmd.Flags().String("url", "", "Reference URL")
	cmd.Flags().String("expired-at", "", "Expiration date (ISO 8601)")
	cmd.Flags().String("broker", "", "Broker ID (omit for global)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostIndexesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostIndexesCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.URL != "" {
		attributes["url"] = opts.URL
	}
	if opts.ExpiredAt != "" {
		attributes["expired-at"] = opts.ExpiredAt
	}

	data := map[string]any{
		"type":       "cost-indexes",
		"attributes": attributes,
	}

	if opts.Broker != "" {
		data["relationships"] = map[string]any{
			"broker": map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			},
		}
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

	body, _, err := client.Post(cmd.Context(), "/v1/cost-indexes", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created cost index %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCostIndexesCreateOptions(cmd *cobra.Command) (doCostIndexesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	url, _ := cmd.Flags().GetString("url")
	expiredAt, _ := cmd.Flags().GetString("expired-at")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostIndexesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		Description: description,
		URL:         url,
		ExpiredAt:   expiredAt,
		Broker:      broker,
	}, nil
}

func buildCostIndexRowFromSingle(resp jsonAPISingleResponse) costIndexRow {
	attrs := resp.Data.Attributes

	row := costIndexRow{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		Description: stringAttr(attrs, "description"),
		URL:         stringAttr(attrs, "url"),
		ExpiredAt:   stringAttr(attrs, "expired-at"),
		IsExpired:   boolAttr(attrs, "is-expired"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
