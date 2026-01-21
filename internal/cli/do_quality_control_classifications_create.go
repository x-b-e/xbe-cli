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

type doQualityControlClassificationsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	Description string
	Broker      string
}

func newDoQualityControlClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new quality control classification",
		Long: `Create a new quality control classification.

Required flags:
  --name    The classification name (required)
  --broker  The broker ID (required)

Optional flags:
  --description  Description of the classification`,
		Example: `  # Create a basic classification
  xbe do quality-control-classifications create --name "Temperature Check" --broker 123

  # Create with description
  xbe do quality-control-classifications create --name "Density Test" --broker 123 --description "Material density verification"

  # Get JSON output
  xbe do quality-control-classifications create --name "Test" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoQualityControlClassificationsCreate,
	}
	initDoQualityControlClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doQualityControlClassificationsCmd.AddCommand(newDoQualityControlClassificationsCreateCmd())
}

func initDoQualityControlClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoQualityControlClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoQualityControlClassificationsCreateOptions(cmd)
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
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "quality-control-classifications",
			"attributes": attributes,
			"relationships": map[string]any{
				"broker": map[string]any{
					"data": map[string]any{
						"type": "brokers",
						"id":   opts.Broker,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/quality-control-classifications", jsonBody)
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

	row := buildQualityControlClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created quality control classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoQualityControlClassificationsCreateOptions(cmd *cobra.Command) (doQualityControlClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doQualityControlClassificationsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		Description: description,
		Broker:      broker,
	}, nil
}

func buildQualityControlClassificationRowFromSingle(resp jsonAPISingleResponse) qualityControlClassificationRow {
	attrs := resp.Data.Attributes

	row := qualityControlClassificationRow{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
