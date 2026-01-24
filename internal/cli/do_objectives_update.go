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

type doObjectivesUpdateOptions struct {
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
	IsAbandoned                                     string
	Owner                                           string
	Parent                                          string
	ParentType                                      string
	ParentID                                        string
	Project                                         string
	SalesResponsiblePerson                          string
}

func newDoObjectivesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an objective",
		Long: `Update an existing objective.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The objective ID (required)

Flags:
  --name                                         Update the name
  --description                                  Update the description
  --start-on                                     Update the start date (YYYY-MM-DD)
  --end-on                                       Update the end date (YYYY-MM-DD)
  --commitment                                   Update the commitment (committed, aspirational)
  --name-summary-explicit                         Update explicit name summary
  --is-template                                  Update template flag (true/false)
  --template-scope                               Update template scope
  --is-generating-objective-stakeholder-classifications  Update generation flag (true/false)
  --is-abandoned                                 Mark as abandoned (true/false)
  --owner                                        Update owner user ID
  --parent                                       Update parent relationship (Objective|ID or KeyResult|ID)
  --parent-type                                  Update parent type (Objective or KeyResult, requires --parent-id)
  --parent-id                                    Update parent ID (requires --parent-type)
  --project                                      Update project ID
  --sales-responsible-person                     Update sales responsible person user ID`,
		Example: `  # Update the name
  xbe do objectives update 123 --name "Updated Objective"

  # Update dates
  xbe do objectives update 123 --start-on 2025-01-01 --end-on 2025-06-30

  # Mark abandoned
  xbe do objectives update 123 --is-abandoned true

  # Output as JSON
  xbe do objectives update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoObjectivesUpdate,
	}
	initDoObjectivesUpdateFlags(cmd)
	return cmd
}

func init() {
	doObjectivesCmd.AddCommand(newDoObjectivesUpdateCmd())
}

func initDoObjectivesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("start-on", "", "New start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "New end date (YYYY-MM-DD)")
	cmd.Flags().String("commitment", "", "New commitment (committed, aspirational)")
	cmd.Flags().String("name-summary-explicit", "", "New explicit name summary")
	cmd.Flags().String("is-template", "", "New template flag (true/false)")
	cmd.Flags().String("template-scope", "", "New template scope")
	cmd.Flags().String("is-generating-objective-stakeholder-classifications", "", "Generate stakeholder classifications (true/false)")
	cmd.Flags().String("is-abandoned", "", "Mark abandoned (true/false)")
	cmd.Flags().String("owner", "", "New owner user ID")
	cmd.Flags().String("parent", "", "Parent relationship (Objective|ID or KeyResult|ID)")
	cmd.Flags().String("parent-type", "", "Parent type (Objective or KeyResult, requires --parent-id)")
	cmd.Flags().String("parent-id", "", "Parent ID (requires --parent-type)")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("sales-responsible-person", "", "Sales responsible person user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectivesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoObjectivesUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("objective id is required")
	}

	if opts.Name == "" && opts.Description == "" && opts.StartOn == "" && opts.EndOn == "" &&
		opts.Commitment == "" && opts.NameSummaryExplicit == "" && opts.IsTemplate == "" &&
		opts.TemplateScope == "" && opts.IsGeneratingObjectiveStakeholderClassifications == "" &&
		opts.IsAbandoned == "" && opts.Owner == "" && opts.Parent == "" && opts.ParentType == "" &&
		opts.ParentID == "" && opts.Project == "" && opts.SalesResponsiblePerson == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	parentType, parentID, err := resolveObjectiveParent(cmd, opts.Parent, opts.ParentType, opts.ParentID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}
	if opts.EndOn != "" {
		attributes["end-on"] = opts.EndOn
	}
	if opts.Commitment != "" {
		attributes["commitment"] = opts.Commitment
	}
	if opts.NameSummaryExplicit != "" {
		attributes["name-summary-explicit"] = opts.NameSummaryExplicit
	}
	if opts.TemplateScope != "" {
		attributes["template-scope"] = opts.TemplateScope
	}
	setBoolAttrIfPresent(attributes, "is-template", opts.IsTemplate)
	setBoolAttrIfPresent(attributes, "is-generating-objective-stakeholder-classifications", opts.IsGeneratingObjectiveStakeholderClassifications)
	setBoolAttrIfPresent(attributes, "is-abandoned", opts.IsAbandoned)

	data := map[string]any{
		"id":         id,
		"type":       "objectives",
		"attributes": attributes,
	}

	relationships := map[string]any{}
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
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/objectives/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated objective %s (%s)\n", result["id"], result["name"])
	return nil
}

func parseDoObjectivesUpdateOptions(cmd *cobra.Command) (doObjectivesUpdateOptions, error) {
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
	isAbandoned, _ := cmd.Flags().GetString("is-abandoned")
	owner, _ := cmd.Flags().GetString("owner")
	parent, _ := cmd.Flags().GetString("parent")
	parentType, _ := cmd.Flags().GetString("parent-type")
	parentID, _ := cmd.Flags().GetString("parent-id")
	project, _ := cmd.Flags().GetString("project")
	salesResponsiblePerson, _ := cmd.Flags().GetString("sales-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectivesUpdateOptions{
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
		IsAbandoned:            isAbandoned,
		Owner:                  owner,
		Parent:                 parent,
		ParentType:             parentType,
		ParentID:               parentID,
		Project:                project,
		SalesResponsiblePerson: salesResponsiblePerson,
	}, nil
}
