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

type doObjectivesCreateOptions struct {
	BaseURL                                         string
	Token                                           string
	JSON                                            bool
	Name                                            string
	Description                                     string
	StartOn                                         string
	EndOn                                           string
	Commitment                                      string
	NameSummaryExplicit                             string
	IsTemplate                                      string
	TemplateScope                                   string
	IsGeneratingObjectiveStakeholderClassifications string
	Owner                                           string
	Organization                                    string
	OrganizationType                                string
	OrganizationID                                  string
	Parent                                          string
	ParentType                                      string
	ParentID                                        string
	Project                                         string
	SalesResponsiblePerson                          string
}

func newDoObjectivesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an objective",
		Long: `Create an objective.

Required flags:
  --name              Objective name (required)

For templates, use:
  --is-template true --template-scope <match_all|organization|project>

For non-templates, specify an organization:
  --organization "Type|ID" or --organization-type/--organization-id

Optional flags:
  --description                                  Description
  --start-on                                      Start date (YYYY-MM-DD)
  --end-on                                        End date (YYYY-MM-DD)
  --commitment                                    Commitment (committed, aspirational)
  --name-summary-explicit                         Explicit name summary
  --is-generating-objective-stakeholder-classifications  Generate stakeholder classifications (true/false)
  --owner                                         Owner user ID
  --parent                                        Parent relationship (Objective|ID or KeyResult|ID)
  --parent-type                                   Parent type (Objective or KeyResult, requires --parent-id)
  --parent-id                                     Parent ID (requires --parent-type)
  --project                                       Project ID
  --sales-responsible-person                      Sales responsible person user ID`,
		Example: `  # Create a non-template objective
  xbe do objectives create --name "Improve On-Time Delivery" --organization "Broker|123"

  # Create a template objective
  xbe do objectives create --name "Customer Experience" --is-template true --template-scope match_all

  # Create with dates and commitment
  xbe do objectives create --name "Reduce rework" --organization "Developer|456" \
    --start-on 2025-01-01 --end-on 2025-06-30 --commitment committed

  # Output as JSON
  xbe do objectives create --name "New Objective" --organization "Broker|123" --json`,
		Args: cobra.NoArgs,
		RunE: runDoObjectivesCreate,
	}
	initDoObjectivesCreateFlags(cmd)
	return cmd
}

func init() {
	doObjectivesCmd.AddCommand(newDoObjectivesCreateCmd())
}

func initDoObjectivesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Objective name (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("commitment", "", "Commitment (committed, aspirational)")
	cmd.Flags().String("name-summary-explicit", "", "Explicit name summary")
	cmd.Flags().String("is-template", "", "Mark as template (true/false)")
	cmd.Flags().String("template-scope", "", "Template scope (match_all, organization, project)")
	cmd.Flags().String("is-generating-objective-stakeholder-classifications", "", "Generate stakeholder classifications (true/false)")
	cmd.Flags().String("owner", "", "Owner user ID")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (e.g., Broker|123)")
	cmd.Flags().String("organization-type", "", "Organization type (optional if --organization is set)")
	cmd.Flags().String("organization-id", "", "Organization ID (optional if --organization is set)")
	cmd.Flags().String("parent", "", "Parent relationship (Objective|ID or KeyResult|ID)")
	cmd.Flags().String("parent-type", "", "Parent type (Objective or KeyResult, requires --parent-id)")
	cmd.Flags().String("parent-id", "", "Parent ID (requires --parent-type)")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("sales-responsible-person", "", "Sales responsible person user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectivesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoObjectivesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Name) == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	orgType, orgID, err := resolveObjectiveOrganization(cmd, opts.Organization, opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	parentType, parentID, err := resolveObjectiveParent(cmd, opts.Parent, opts.ParentType, opts.ParentID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	isTemplate := strings.EqualFold(strings.TrimSpace(opts.IsTemplate), "true")
	if isTemplate {
		if strings.TrimSpace(opts.TemplateScope) == "" {
			err := fmt.Errorf("--template-scope is required when --is-template true")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if orgType != "" || orgID != "" {
			err := fmt.Errorf("--organization cannot be used when --is-template true")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if opts.Owner != "" {
			err := fmt.Errorf("--owner cannot be used when --is-template true")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if opts.SalesResponsiblePerson != "" {
			err := fmt.Errorf("--sales-responsible-person cannot be used when --is-template true")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	} else if strings.TrimSpace(opts.TemplateScope) != "" {
		err := fmt.Errorf("--template-scope requires --is-template true")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !isTemplate && orgType == "" && orgID == "" {
		err := fmt.Errorf("--organization is required for non-template objectives (or set --is-template true)")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}
	setStringAttrIfPresent(attributes, "description", opts.Description)
	setStringAttrIfPresent(attributes, "start-on", opts.StartOn)
	setStringAttrIfPresent(attributes, "end-on", opts.EndOn)
	setStringAttrIfPresent(attributes, "commitment", opts.Commitment)
	setStringAttrIfPresent(attributes, "name-summary-explicit", opts.NameSummaryExplicit)
	setStringAttrIfPresent(attributes, "template-scope", opts.TemplateScope)
	setBoolAttrIfPresent(attributes, "is-template", opts.IsTemplate)
	setBoolAttrIfPresent(attributes, "is-generating-objective-stakeholder-classifications", opts.IsGeneratingObjectiveStakeholderClassifications)

	relationships := map[string]any{}
	if orgType != "" && orgID != "" {
		relationships["organization"] = map[string]any{
			"data": map[string]string{
				"type": orgType,
				"id":   orgID,
			},
		}
	}
	if opts.Owner != "" {
		relationships["owner"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.Owner,
			},
		}
	}
	if parentType != "" && parentID != "" {
		relationships["parent"] = map[string]any{
			"data": map[string]string{
				"type": parentType,
				"id":   parentID,
			},
		}
	}
	if opts.Project != "" {
		relationships["project"] = map[string]any{
			"data": map[string]string{
				"type": "projects",
				"id":   opts.Project,
			},
		}
	}
	if opts.SalesResponsiblePerson != "" {
		relationships["sales-responsible-person"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.SalesResponsiblePerson,
			},
		}
	}

	requestData := map[string]any{
		"type":       "objectives",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/objectives", jsonBody)
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

	result := map[string]any{
		"id":   resp.Data.ID,
		"name": stringAttr(resp.Data.Attributes, "name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created objective %s (%s)\n", result["id"], result["name"])
	return nil
}

func parseDoObjectivesCreateOptions(cmd *cobra.Command) (doObjectivesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	commitment, _ := cmd.Flags().GetString("commitment")
	nameSummaryExplicit, _ := cmd.Flags().GetString("name-summary-explicit")
	isTemplate, _ := cmd.Flags().GetString("is-template")
	templateScope, _ := cmd.Flags().GetString("template-scope")
	isGeneratingObjectiveStakeholderClassifications, _ := cmd.Flags().GetString("is-generating-objective-stakeholder-classifications")
	owner, _ := cmd.Flags().GetString("owner")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	parent, _ := cmd.Flags().GetString("parent")
	parentType, _ := cmd.Flags().GetString("parent-type")
	parentID, _ := cmd.Flags().GetString("parent-id")
	project, _ := cmd.Flags().GetString("project")
	salesResponsiblePerson, _ := cmd.Flags().GetString("sales-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectivesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Name:                name,
		Description:         description,
		StartOn:             startOn,
		EndOn:               endOn,
		Commitment:          commitment,
		NameSummaryExplicit: nameSummaryExplicit,
		IsTemplate:          isTemplate,
		TemplateScope:       templateScope,
		IsGeneratingObjectiveStakeholderClassifications: isGeneratingObjectiveStakeholderClassifications,
		Owner:                  owner,
		Organization:           organization,
		OrganizationType:       organizationType,
		OrganizationID:         organizationID,
		Parent:                 parent,
		ParentType:             parentType,
		ParentID:               parentID,
		Project:                project,
		SalesResponsiblePerson: salesResponsiblePerson,
	}, nil
}
