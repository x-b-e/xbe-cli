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

type doEquipmentClassificationsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Name               string
	Abbreviation       string
	MobilizationMethod string
	Parent             string
}

func newDoEquipmentClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new equipment classification",
		Long: `Create a new equipment classification.

Required flags:
  --name          The equipment classification name (required)

Optional flags:
  --abbreviation         Short code for the classification
  --mobilization-method  How equipment is mobilized (crew, heavy_equipment_transport, itself, trailer)
  --parent               Parent classification ID for hierarchical organization`,
		Example: `  # Create a basic equipment classification
  xbe do equipment-classifications create --name "Paver"

  # Create with abbreviation and mobilization method
  xbe do equipment-classifications create --name "Paver" --abbreviation "paver" --mobilization-method lowboy

  # Create a child classification
  xbe do equipment-classifications create --name "Asphalt Paver" --abbreviation "asph-paver" --parent 123

  # Get JSON output
  xbe do equipment-classifications create --name "Paver" --json`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentClassificationsCreate,
	}
	initDoEquipmentClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentClassificationsCmd.AddCommand(newDoEquipmentClassificationsCreateCmd())
}

func initDoEquipmentClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Equipment classification name (required)")
	cmd.Flags().String("abbreviation", "", "Short code for the classification")
	cmd.Flags().String("mobilization-method", "", "How equipment is mobilized (crew, heavy_equipment_transport, itself, trailer)")
	cmd.Flags().String("parent", "", "Parent classification ID for hierarchical organization")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentClassificationsCreateOptions(cmd)
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

	// Build attributes
	attributes := map[string]any{
		"name": opts.Name,
	}
	if opts.Abbreviation != "" {
		attributes["abbreviation"] = opts.Abbreviation
	}
	if opts.MobilizationMethod != "" {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}

	// Build request data
	data := map[string]any{
		"type":       "equipment-classifications",
		"attributes": attributes,
	}

	// Add parent relationship if specified
	if opts.Parent != "" {
		data["relationships"] = map[string]any{
			"parent": map[string]any{
				"data": map[string]string{
					"type": "equipment-classifications",
					"id":   opts.Parent,
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-classifications", jsonBody)
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

	row := buildEquipmentClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoEquipmentClassificationsCreateOptions(cmd *cobra.Command) (doEquipmentClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentClassificationsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		Abbreviation:       abbreviation,
		MobilizationMethod: mobilizationMethod,
		Parent:             parent,
	}, nil
}

func buildEquipmentClassificationRowFromSingle(resp jsonAPISingleResponse) equipmentClassificationRow {
	attrs := resp.Data.Attributes

	row := equipmentClassificationRow{
		ID:                 resp.Data.ID,
		Name:               stringAttr(attrs, "name"),
		Abbreviation:       stringAttr(attrs, "abbreviation"),
		MobilizationMethod: stringAttr(attrs, "mobilization-method"),
	}

	// Get parent ID from relationships
	if rel, ok := resp.Data.Relationships["parent"]; ok && rel.Data != nil {
		row.ParentID = rel.Data.ID
	}

	return row
}
