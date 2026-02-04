package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type objectivesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type objectiveDetails struct {
	ID                                              string   `json:"id"`
	Name                                            string   `json:"name,omitempty"`
	Description                                     string   `json:"description,omitempty"`
	Status                                          string   `json:"status,omitempty"`
	StartOn                                         string   `json:"start_on,omitempty"`
	EndOn                                           string   `json:"end_on,omitempty"`
	Commitment                                      string   `json:"commitment,omitempty"`
	NameSummary                                     string   `json:"name_summary,omitempty"`
	NameSummaryExplicit                             string   `json:"name_summary_explicit,omitempty"`
	NameSummaryImplicit                             string   `json:"name_summary_implicit,omitempty"`
	IsTemplate                                      bool     `json:"is_template"`
	TemplateScope                                   string   `json:"template_scope,omitempty"`
	IsGeneratingObjectiveStakeholderClassifications bool     `json:"is_generating_objective_stakeholder_classifications"`
	IsAbandoned                                     bool     `json:"is_abandoned"`
	Slug                                            string   `json:"slug,omitempty"`
	CompletionPercentageCalculated                  any      `json:"completion_percentage_calculated,omitempty"`
	OwnerID                                         string   `json:"owner_id,omitempty"`
	OwnerName                                       string   `json:"owner_name,omitempty"`
	OrganizationID                                  string   `json:"organization_id,omitempty"`
	OrganizationType                                string   `json:"organization_type,omitempty"`
	OrganizationName                                string   `json:"organization_name,omitempty"`
	ParentID                                        string   `json:"parent_id,omitempty"`
	ParentType                                      string   `json:"parent_type,omitempty"`
	ParentName                                      string   `json:"parent_name,omitempty"`
	ProjectID                                       string   `json:"project_id,omitempty"`
	ProjectName                                     string   `json:"project_name,omitempty"`
	SalesResponsiblePersonID                        string   `json:"sales_responsible_person_id,omitempty"`
	SalesResponsiblePersonName                      string   `json:"sales_responsible_person_name,omitempty"`
	KeyResultIDs                                    []string `json:"key_result_ids,omitempty"`
	ChildObjectiveIDs                               []string `json:"child_objective_ids,omitempty"`
	ObjectiveStakeholderClassificationIDs           []string `json:"objective_stakeholder_classification_ids,omitempty"`
	LatestObjectiveStatusPostID                     string   `json:"latest_objective_status_post_id,omitempty"`
	LatestObjectiveStatusPostSummary                string   `json:"latest_objective_status_post_summary,omitempty"`
}

func newObjectivesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show objective details",
		Long: `Show the full details of an objective.

Output Sections:
  Core fields (name, status, dates)
  Template settings and computed fields
  Relationships (owner, organization, project, parent, key results)

Arguments:
  <id>    The objective ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an objective
  xbe view objectives show 123

  # Output as JSON
  xbe view objectives show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runObjectivesShow,
	}
	initObjectivesShowFlags(cmd)
	return cmd
}

func init() {
	objectivesCmd.AddCommand(newObjectivesShowCmd())
}

func initObjectivesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectivesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseObjectivesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("objective id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "owner,organization,parent,project,sales-responsible-person,latest-objective-status-post")
	query.Set("fields[users]", "name")
	query.Set("fields[projects]", "name")
	query.Set("fields[posts]", "short-text-content")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/objectives/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildObjectiveDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderObjectiveDetails(cmd, details)
}

func parseObjectivesShowOptions(cmd *cobra.Command) (objectivesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectivesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildObjectiveDetails(resp jsonAPISingleResponse) objectiveDetails {
	included := indexIncludedResources(resp.Included)
	attrs := resp.Data.Attributes

	details := objectiveDetails{
		ID:                  resp.Data.ID,
		Name:                strings.TrimSpace(stringAttr(attrs, "name")),
		Description:         strings.TrimSpace(stringAttr(attrs, "description")),
		Status:              stringAttr(attrs, "status"),
		StartOn:             formatDate(stringAttr(attrs, "start-on")),
		EndOn:               formatDate(stringAttr(attrs, "end-on")),
		Commitment:          stringAttr(attrs, "commitment"),
		NameSummary:         stringAttr(attrs, "name-summary"),
		NameSummaryExplicit: stringAttr(attrs, "name-summary-explicit"),
		NameSummaryImplicit: stringAttr(attrs, "name-summary-implicit"),
		IsTemplate:          boolAttr(attrs, "is-template"),
		TemplateScope:       stringAttr(attrs, "template-scope"),
		IsGeneratingObjectiveStakeholderClassifications: boolAttr(attrs, "is-generating-objective-stakeholder-classifications"),
		IsAbandoned:                    boolAttr(attrs, "is-abandoned"),
		Slug:                           stringAttr(attrs, "slug"),
		CompletionPercentageCalculated: attrs["completion-percentage-calculated"],
	}

	if rel, ok := resp.Data.Relationships["owner"]; ok && rel.Data != nil {
		details.OwnerID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OwnerName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationID = rel.Data.ID
		details.OrganizationType = rel.Data.Type
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OrganizationName = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["parent"]; ok && rel.Data != nil {
		details.ParentID = rel.Data.ID
		details.ParentType = rel.Data.Type
		if parent, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			switch rel.Data.Type {
			case "objectives":
				details.ParentName = strings.TrimSpace(stringAttr(parent.Attributes, "name"))
			case "key-results":
				details.ParentName = strings.TrimSpace(stringAttr(parent.Attributes, "title"))
			default:
				details.ParentName = firstNonEmpty(
					strings.TrimSpace(stringAttr(parent.Attributes, "name")),
					strings.TrimSpace(stringAttr(parent.Attributes, "title")),
				)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if proj, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = strings.TrimSpace(stringAttr(proj.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["sales-responsible-person"]; ok && rel.Data != nil {
		details.SalesResponsiblePersonID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SalesResponsiblePersonName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["key-results"]; ok && rel.raw != nil {
		details.KeyResultIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["children"]; ok && rel.raw != nil {
		details.ChildObjectiveIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["objective-stakeholder-classifications"]; ok && rel.raw != nil {
		details.ObjectiveStakeholderClassificationIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["latest-objective-status-post"]; ok && rel.Data != nil {
		details.LatestObjectiveStatusPostID = rel.Data.ID
		if post, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LatestObjectiveStatusPostSummary = strings.TrimSpace(stringAttr(post.Attributes, "short-text-content"))
		}
	}

	return details
}

func renderObjectiveDetails(cmd *cobra.Command, d objectiveDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", d.Name)
	}
	if d.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", d.Status)
	}
	if d.Commitment != "" {
		fmt.Fprintf(out, "Commitment: %s\n", d.Commitment)
	}
	if d.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", d.StartOn)
	}
	if d.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", d.EndOn)
	}
	if d.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", d.Description)
	}
	if d.NameSummary != "" {
		fmt.Fprintf(out, "Name Summary: %s\n", d.NameSummary)
	}
	if d.NameSummaryExplicit != "" {
		fmt.Fprintf(out, "Name Summary Explicit: %s\n", d.NameSummaryExplicit)
	}
	if d.NameSummaryImplicit != "" {
		fmt.Fprintf(out, "Name Summary Implicit: %s\n", d.NameSummaryImplicit)
	}
	if d.Slug != "" {
		fmt.Fprintf(out, "Slug: %s\n", d.Slug)
	}
	fmt.Fprintf(out, "Is Template: %t\n", d.IsTemplate)
	if d.TemplateScope != "" {
		fmt.Fprintf(out, "Template Scope: %s\n", d.TemplateScope)
	}
	fmt.Fprintf(out, "Is Generating Stakeholder Classifications: %t\n", d.IsGeneratingObjectiveStakeholderClassifications)
	fmt.Fprintf(out, "Is Abandoned: %t\n", d.IsAbandoned)
	if d.CompletionPercentageCalculated != nil {
		fmt.Fprintf(out, "Completion (Calculated): %s\n", formatAnyValue(d.CompletionPercentageCalculated))
	}

	hasRelationships := d.OwnerID != "" || d.OrganizationID != "" || d.ParentID != "" || d.ProjectID != "" ||
		d.SalesResponsiblePersonID != "" || len(d.KeyResultIDs) > 0 || len(d.ChildObjectiveIDs) > 0 ||
		len(d.ObjectiveStakeholderClassificationIDs) > 0 || d.LatestObjectiveStatusPostID != ""

	if hasRelationships {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Relationships:")
		fmt.Fprintln(out, strings.Repeat("-", 40))

		if d.OwnerID != "" {
			fmt.Fprintf(out, "  Owner: %s\n", formatRelated(d.OwnerName, d.OwnerID))
		}
		if d.OrganizationID != "" || d.OrganizationType != "" {
			orgLabel := formatRelated(d.OrganizationName, formatPolymorphic(d.OrganizationType, d.OrganizationID))
			fmt.Fprintf(out, "  Organization: %s\n", orgLabel)
		}
		if d.ParentID != "" || d.ParentType != "" {
			parentLabel := formatRelated(d.ParentName, formatPolymorphic(d.ParentType, d.ParentID))
			fmt.Fprintf(out, "  Parent: %s\n", parentLabel)
		}
		if d.ProjectID != "" {
			fmt.Fprintf(out, "  Project: %s\n", formatRelated(d.ProjectName, d.ProjectID))
		}
		if d.SalesResponsiblePersonID != "" {
			fmt.Fprintf(out, "  Sales Responsible Person: %s\n", formatRelated(d.SalesResponsiblePersonName, d.SalesResponsiblePersonID))
		}
		if d.LatestObjectiveStatusPostID != "" {
			fmt.Fprintf(out, "  Latest Status Post: %s\n", formatRelated(d.LatestObjectiveStatusPostSummary, d.LatestObjectiveStatusPostID))
		}
		if len(d.KeyResultIDs) > 0 {
			fmt.Fprintf(out, "  Key Results: %s\n", strings.Join(d.KeyResultIDs, ", "))
		}
		if len(d.ChildObjectiveIDs) > 0 {
			fmt.Fprintf(out, "  Child Objectives: %s\n", strings.Join(d.ChildObjectiveIDs, ", "))
		}
		if len(d.ObjectiveStakeholderClassificationIDs) > 0 {
			fmt.Fprintf(out, "  Objective Stakeholder Classifications: %s\n", strings.Join(d.ObjectiveStakeholderClassificationIDs, ", "))
		}
	}

	return nil
}
