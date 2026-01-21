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

type doProjectCostClassificationsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Name    string
	Broker  string
	Parent  string
}

func newDoProjectCostClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project cost classification",
		Long: `Create a new project cost classification.

Required flags:
  --name    The classification name (required)
  --broker  The broker ID (required)

Optional flags:
  --parent  Parent classification ID (for hierarchical structure)`,
		Example: `  # Create a root classification
  xbe do project-cost-classifications create --name "Labor" --broker 123

  # Create a child classification
  xbe do project-cost-classifications create --name "Skilled Labor" --broker 123 --parent 456

  # Get JSON output
  xbe do project-cost-classifications create --name "Materials" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectCostClassificationsCreate,
	}
	initDoProjectCostClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectCostClassificationsCmd.AddCommand(newDoProjectCostClassificationsCreateCmd())
}

func initDoProjectCostClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("parent", "", "Parent classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectCostClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectCostClassificationsCreateOptions(cmd)
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

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if opts.Parent != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": "project-cost-classifications",
				"id":   opts.Parent,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-cost-classifications",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-cost-classifications", jsonBody)
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

	row := buildProjectCostClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project cost classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectCostClassificationsCreateOptions(cmd *cobra.Command) (doProjectCostClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCostClassificationsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Name:    name,
		Broker:  broker,
		Parent:  parent,
	}, nil
}

func buildProjectCostClassificationRowFromSingle(resp jsonAPISingleResponse) projectCostClassificationRow {
	attrs := resp.Data.Attributes

	row := projectCostClassificationRow{
		ID:   resp.Data.ID,
		Name: stringAttr(attrs, "name"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["parent"]; ok && rel.Data != nil {
		row.ParentID = rel.Data.ID
	}

	return row
}
