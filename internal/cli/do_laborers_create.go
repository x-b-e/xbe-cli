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

type doLaborersCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	IsActive              bool
	MobilizationMethod    string
	GroupName             string
	ColorHex              string
	LaborClassificationID string
	UserID                string
	OrganizationType      string
	OrganizationID        string
	CraftClassID          string
}

func newDoLaborersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new laborer",
		Long: `Create a new laborer.

Required flags:
  --labor-classification   Labor classification ID (required)
  --user                   User ID (required)
  --organization-type      Organization type (e.g., brokers, customers) (required)
  --organization-id        Organization ID (required)

Optional flags:
  --is-active              Whether laborer is active (default true)
  --mobilization-method    Mobilization method
  --group-name             Group name
  --color-hex              Color hex code
  --craft-class            Craft class ID`,
		Example: `  # Create laborer with required fields
  xbe do laborers create \
    --labor-classification 123 \
    --user 456 \
    --organization-type brokers \
    --organization-id 789

  # Create laborer with optional fields
  xbe do laborers create \
    --labor-classification 123 \
    --user 456 \
    --organization-type brokers \
    --organization-id 789 \
    --group-name "Crew A" \
    --color-hex "#FF0000"`,
		Args: cobra.NoArgs,
		RunE: runDoLaborersCreate,
	}
	initDoLaborersCreateFlags(cmd)
	return cmd
}

func init() {
	doLaborersCmd.AddCommand(newDoLaborersCreateCmd())
}

func initDoLaborersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-active", true, "Whether laborer is active")
	cmd.Flags().String("mobilization-method", "", "Mobilization method")
	cmd.Flags().String("group-name", "", "Group name")
	cmd.Flags().String("color-hex", "", "Color hex code")
	cmd.Flags().String("labor-classification", "", "Labor classification ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("organization-type", "", "Organization type (e.g., brokers, customers) (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("craft-class", "", "Craft class ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaborersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaborersCreateOptions(cmd)
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

	if opts.LaborClassificationID == "" {
		err := fmt.Errorf("--labor-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.UserID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"is-active": opts.IsActive,
	}

	if opts.MobilizationMethod != "" {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}
	if opts.GroupName != "" {
		attributes["group-name"] = opts.GroupName
	}
	if opts.ColorHex != "" {
		attributes["color-hex"] = opts.ColorHex
	}

	relationships := map[string]any{
		"labor-classification": map[string]any{
			"data": map[string]any{
				"type": "labor-classifications",
				"id":   opts.LaborClassificationID,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	if opts.CraftClassID != "" {
		relationships["craft-class"] = map[string]any{
			"data": map[string]any{
				"type": "craft-classes",
				"id":   opts.CraftClassID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "laborers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/laborers", jsonBody)
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

	row := buildLaborerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created laborer %s\n", row.ID)
	return nil
}

func parseDoLaborersCreateOptions(cmd *cobra.Command) (doLaborersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isActive, _ := cmd.Flags().GetBool("is-active")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	groupName, _ := cmd.Flags().GetString("group-name")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	laborClassificationID, _ := cmd.Flags().GetString("labor-classification")
	userID, _ := cmd.Flags().GetString("user")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	craftClassID, _ := cmd.Flags().GetString("craft-class")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLaborersCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		IsActive:              isActive,
		MobilizationMethod:    mobilizationMethod,
		GroupName:             groupName,
		ColorHex:              colorHex,
		LaborClassificationID: laborClassificationID,
		UserID:                userID,
		OrganizationType:      organizationType,
		OrganizationID:        organizationID,
		CraftClassID:          craftClassID,
	}, nil
}

func buildLaborerRowFromSingle(resp jsonAPISingleResponse) laborerRow {
	attrs := resp.Data.Attributes

	row := laborerRow{
		ID:       resp.Data.ID,
		Nickname: stringAttr(attrs, "nickname"),
		IsActive: boolAttr(attrs, "is-active"),
	}

	if rel, ok := resp.Data.Relationships["labor-classification"]; ok && rel.Data != nil {
		row.LaborClassificationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}
