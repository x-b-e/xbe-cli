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

type doEquipmentMovementRequirementLocationsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Name      string
	Latitude  string
	Longitude string
	BrokerID  string
}

func newDoEquipmentMovementRequirementLocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement requirement location",
		Long: `Create an equipment movement requirement location.

Required flags:
  --broker      Broker ID (required)
  --latitude    Latitude coordinate (required)
  --longitude   Longitude coordinate (required)

Optional flags:
  --name        Location name`,
		Example: `  # Create a location
  xbe do equipment-movement-requirement-locations create \
    --broker 123 \
    --latitude 37.7749 \
    --longitude -122.4194 \
    --name "Main Yard"`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementRequirementLocationsCreate,
	}
	initDoEquipmentMovementRequirementLocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementRequirementLocationsCmd.AddCommand(newDoEquipmentMovementRequirementLocationsCreateCmd())
}

func initDoEquipmentMovementRequirementLocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("latitude", "", "Latitude coordinate")
	cmd.Flags().String("longitude", "", "Longitude coordinate")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("latitude")
	_ = cmd.MarkFlagRequired("longitude")
}

func runDoEquipmentMovementRequirementLocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementRequirementLocationsCreateOptions(cmd)
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

	if opts.BrokerID == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Latitude == "" {
		err := fmt.Errorf("--latitude is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Longitude == "" {
		err := fmt.Errorf("--longitude is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"latitude":  opts.Latitude,
		"longitude": opts.Longitude,
	}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-requirement-locations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-requirement-locations", jsonBody)
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

	row := buildEquipmentMovementRequirementLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement requirement location %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementRequirementLocationsCreateOptions(cmd *cobra.Command) (doEquipmentMovementRequirementLocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	brokerID, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementRequirementLocationsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Name:      name,
		Latitude:  latitude,
		Longitude: longitude,
		BrokerID:  brokerID,
	}, nil
}
