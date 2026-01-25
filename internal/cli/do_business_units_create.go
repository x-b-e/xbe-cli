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

type doBusinessUnitsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Name       string
	ExternalID string
	Broker     string
	Parent     string
}

func newDoBusinessUnitsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new business unit",
		Long: `Create a new business unit.

Required flags:
  --name      The business unit name (required)
  --broker    The broker ID (required)

Optional flags:
  --external-id    External identifier
  --parent         Parent business unit ID`,
		Example: `  # Create a business unit
  xbe do business-units create --name "Paving Division" --broker 123

  # Create with external ID
  xbe do business-units create --name "Concrete" --broker 123 --external-id "BU-001"

  # Create as child of another business unit
  xbe do business-units create --name "Sub-Division" --broker 123 --parent 456

  # Get JSON output
  xbe do business-units create --name "New Unit" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoBusinessUnitsCreate,
	}
	initDoBusinessUnitsCreateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitsCmd.AddCommand(newDoBusinessUnitsCreateCmd())
}

func initDoBusinessUnitsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Business unit name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("external-id", "", "External identifier")
	cmd.Flags().String("parent", "", "Parent business unit ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBusinessUnitsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require broker
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"company-name": opts.Name,
	}
	if opts.ExternalID != "" {
		attributes["external-id"] = opts.ExternalID
	}

	// Build relationships
	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]string{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}
	if opts.Parent != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]string{
				"type": "business-units",
				"id":   opts.Parent,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "business-units",
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

	body, _, err := client.Post(cmd.Context(), "/v1/business-units", jsonBody)
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

	row := businessUnitRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "company-name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created business unit %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoBusinessUnitsCreateOptions(cmd *cobra.Command) (doBusinessUnitsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	externalID, _ := cmd.Flags().GetString("external-id")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Name:       name,
		Broker:     broker,
		ExternalID: externalID,
		Parent:     parent,
	}, nil
}
