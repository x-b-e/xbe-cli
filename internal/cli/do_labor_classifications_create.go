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

type doLaborClassificationsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Name               string
	Abbreviation       string
	MobilizationMethod string
	IsTimeCardApprover bool
	AllowColors        bool
	IsManager          bool
	CanManageProjects  bool
}

func newDoLaborClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new labor classification",
		Long: `Create a new labor classification.

Required flags:
  --name          The labor classification name (required)

Optional flags:
  --abbreviation         Short code for the classification
  --mobilization-method  How workers are mobilized
  --is-time-card-approver  Can approve time cards
  --allow-colors         Allow color assignments
  --is-manager           Is a manager role
  --can-manage-projects  Can manage projects`,
		Example: `  # Create a basic labor classification
  xbe do labor-classifications create --name "Raker"

  # Create with abbreviation
  xbe do labor-classifications create --name "Raker" --abbreviation "raker"

  # Create a manager classification
  xbe do labor-classifications create --name "Foreman" --abbreviation "foreman" --is-manager

  # Get JSON output
  xbe do labor-classifications create --name "Raker" --json`,
		Args: cobra.NoArgs,
		RunE: runDoLaborClassificationsCreate,
	}
	initDoLaborClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doLaborClassificationsCmd.AddCommand(newDoLaborClassificationsCreateCmd())
}

func initDoLaborClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Labor classification name (required)")
	cmd.Flags().String("abbreviation", "", "Short code for the classification")
	cmd.Flags().String("mobilization-method", "", "How workers are mobilized")
	cmd.Flags().Bool("is-time-card-approver", false, "Can approve time cards")
	cmd.Flags().Bool("allow-colors", false, "Allow color assignments")
	cmd.Flags().Bool("is-manager", false, "Is a manager role")
	cmd.Flags().Bool("can-manage-projects", false, "Can manage projects")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaborClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaborClassificationsCreateOptions(cmd)
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
	if cmd.Flags().Changed("is-time-card-approver") {
		attributes["is-time-card-approver"] = opts.IsTimeCardApprover
	}
	if cmd.Flags().Changed("allow-colors") {
		attributes["allow-colors"] = opts.AllowColors
	}
	if cmd.Flags().Changed("is-manager") {
		attributes["is-manager"] = opts.IsManager
	}
	if cmd.Flags().Changed("can-manage-projects") {
		attributes["can-manage-projects"] = opts.CanManageProjects
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "labor-classifications",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/labor-classifications", jsonBody)
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

	row := buildLaborClassificationRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created labor classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoLaborClassificationsCreateOptions(cmd *cobra.Command) (doLaborClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	isTimeCardApprover, _ := cmd.Flags().GetBool("is-time-card-approver")
	allowColors, _ := cmd.Flags().GetBool("allow-colors")
	isManager, _ := cmd.Flags().GetBool("is-manager")
	canManageProjects, _ := cmd.Flags().GetBool("can-manage-projects")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLaborClassificationsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		Abbreviation:       abbreviation,
		MobilizationMethod: mobilizationMethod,
		IsTimeCardApprover: isTimeCardApprover,
		AllowColors:        allowColors,
		IsManager:          isManager,
		CanManageProjects:  canManageProjects,
	}, nil
}

type laborClassificationRow struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Abbreviation       string `json:"abbreviation,omitempty"`
	MobilizationMethod string `json:"mobilization_method,omitempty"`
	IsTimeCardApprover bool   `json:"is_time_card_approver"`
	AllowColors        bool   `json:"allow_colors"`
	IsManager          bool   `json:"is_manager"`
	CanManageProjects  bool   `json:"can_manage_projects"`
}

func buildLaborClassificationRow(resp jsonAPISingleResponse) laborClassificationRow {
	attrs := resp.Data.Attributes

	return laborClassificationRow{
		ID:                 resp.Data.ID,
		Name:               stringAttr(attrs, "name"),
		Abbreviation:       stringAttr(attrs, "abbreviation"),
		MobilizationMethod: stringAttr(attrs, "mobilization-method"),
		IsTimeCardApprover: boolAttr(attrs, "is-time-card-approver"),
		AllowColors:        boolAttr(attrs, "allow-colors"),
		IsManager:          boolAttr(attrs, "is-manager"),
		CanManageProjects:  boolAttr(attrs, "can-manage-projects"),
	}
}
