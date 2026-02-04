package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanMaterialSiteChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanMaterialSiteChangeDetails struct {
	ID                     string `json:"id"`
	JobProductionPlan      string `json:"job_production_plan,omitempty"`
	JobProductionPlanID    string `json:"job_production_plan_id,omitempty"`
	OldMaterialSite        string `json:"old_material_site,omitempty"`
	OldMaterialSiteID      string `json:"old_material_site_id,omitempty"`
	NewMaterialSite        string `json:"new_material_site,omitempty"`
	NewMaterialSiteID      string `json:"new_material_site_id,omitempty"`
	OldMaterialType        string `json:"old_material_type,omitempty"`
	OldMaterialTypeID      string `json:"old_material_type_id,omitempty"`
	NewMaterialType        string `json:"new_material_type,omitempty"`
	NewMaterialTypeID      string `json:"new_material_type_id,omitempty"`
	NewMaterialMixDesign   string `json:"new_material_mix_design,omitempty"`
	NewMaterialMixDesignID string `json:"new_material_mix_design_id,omitempty"`
	CreatedBy              string `json:"created_by,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
}

func newJobProductionPlanMaterialSiteChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan material site change details",
		Long: `Show the full details of a job production plan material site change.

Output Fields:
  ID                     Material site change identifier
  Job Production Plan    Job production plan name/number
  Old Material Site      Old material site
  New Material Site      New material site
  Old Material Type      Old material type (if applicable)
  New Material Type      New material type (if applicable)
  New Material Mix Design New material mix design (if applicable)
  Created By             User who created the change
  Created                Created timestamp
  Updated                Updated timestamp

Arguments:
  <id>  The material site change ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show material site change details
  xbe view job-production-plan-material-site-changes show 123

  # Show as JSON
  xbe view job-production-plan-material-site-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanMaterialSiteChangesShow,
	}
	initJobProductionPlanMaterialSiteChangesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialSiteChangesCmd.AddCommand(newJobProductionPlanMaterialSiteChangesShowCmd())
}

func initJobProductionPlanMaterialSiteChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialSiteChangesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanMaterialSiteChangesShowOptions(cmd)
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
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("material site change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-material-site-changes]", "created-at,updated-at,job-production-plan,old-material-site,new-material-site,old-material-type,new-material-type,new-material-mix-design,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "display-name,name,fully-qualified-name")
	query.Set("fields[material-mix-designs]", "description,mix")
	query.Set("fields[users]", "name")
	query.Set("include", "job-production-plan,old-material-site,new-material-site,old-material-type,new-material-type,new-material-mix-design,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-site-changes/"+id, query)
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

	details := buildJobProductionPlanMaterialSiteChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanMaterialSiteChangeDetails(cmd, details)
}

func parseJobProductionPlanMaterialSiteChangesShowOptions(cmd *cobra.Command) (jobProductionPlanMaterialSiteChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanMaterialSiteChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanMaterialSiteChangeDetails(resp jsonAPISingleResponse) jobProductionPlanMaterialSiteChangeDetails {
	attrs := resp.Data.Attributes
	details := jobProductionPlanMaterialSiteChangeDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	jppType := ""
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}
	oldSiteType := ""
	if rel, ok := resp.Data.Relationships["old-material-site"]; ok && rel.Data != nil {
		details.OldMaterialSiteID = rel.Data.ID
		oldSiteType = rel.Data.Type
	}
	newSiteType := ""
	if rel, ok := resp.Data.Relationships["new-material-site"]; ok && rel.Data != nil {
		details.NewMaterialSiteID = rel.Data.ID
		newSiteType = rel.Data.Type
	}
	oldTypeType := ""
	if rel, ok := resp.Data.Relationships["old-material-type"]; ok && rel.Data != nil {
		details.OldMaterialTypeID = rel.Data.ID
		oldTypeType = rel.Data.Type
	}
	newTypeType := ""
	if rel, ok := resp.Data.Relationships["new-material-type"]; ok && rel.Data != nil {
		details.NewMaterialTypeID = rel.Data.ID
		newTypeType = rel.Data.Type
	}
	mixDesignType := ""
	if rel, ok := resp.Data.Relationships["new-material-mix-design"]; ok && rel.Data != nil {
		details.NewMaterialMixDesignID = rel.Data.ID
		mixDesignType = rel.Data.Type
	}
	createdByType := ""
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	if details.JobProductionPlanID != "" && jppType != "" {
		if jpp, ok := included[resourceKey(jppType, details.JobProductionPlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(jpp.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(jpp.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if details.OldMaterialSiteID != "" && oldSiteType != "" {
		if site, ok := included[resourceKey(oldSiteType, details.OldMaterialSiteID)]; ok {
			details.OldMaterialSite = strings.TrimSpace(stringAttr(site.Attributes, "name"))
		}
	}

	if details.NewMaterialSiteID != "" && newSiteType != "" {
		if site, ok := included[resourceKey(newSiteType, details.NewMaterialSiteID)]; ok {
			details.NewMaterialSite = strings.TrimSpace(stringAttr(site.Attributes, "name"))
		}
	}

	if details.OldMaterialTypeID != "" && oldTypeType != "" {
		if materialType, ok := included[resourceKey(oldTypeType, details.OldMaterialTypeID)]; ok {
			details.OldMaterialType = firstNonEmpty(
				strings.TrimSpace(stringAttr(materialType.Attributes, "display-name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "fully-qualified-name")),
			)
		}
	}

	if details.NewMaterialTypeID != "" && newTypeType != "" {
		if materialType, ok := included[resourceKey(newTypeType, details.NewMaterialTypeID)]; ok {
			details.NewMaterialType = firstNonEmpty(
				strings.TrimSpace(stringAttr(materialType.Attributes, "display-name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "name")),
				strings.TrimSpace(stringAttr(materialType.Attributes, "fully-qualified-name")),
			)
		}
	}

	if details.NewMaterialMixDesignID != "" && mixDesignType != "" {
		if mixDesign, ok := included[resourceKey(mixDesignType, details.NewMaterialMixDesignID)]; ok {
			details.NewMaterialMixDesign = firstNonEmpty(
				strings.TrimSpace(stringAttr(mixDesign.Attributes, "description")),
				strings.TrimSpace(stringAttr(mixDesign.Attributes, "mix")),
			)
		}
	}

	if details.CreatedByID != "" && createdByType != "" {
		if user, ok := included[resourceKey(createdByType, details.CreatedByID)]; ok {
			details.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return details
}

func renderJobProductionPlanMaterialSiteChangeDetails(cmd *cobra.Command, details jobProductionPlanMaterialSiteChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	renderMaterialSiteChangeRelation(out, "Job Production Plan", details.JobProductionPlan, details.JobProductionPlanID)
	renderMaterialSiteChangeRelation(out, "Old Material Site", details.OldMaterialSite, details.OldMaterialSiteID)
	renderMaterialSiteChangeRelation(out, "New Material Site", details.NewMaterialSite, details.NewMaterialSiteID)
	renderMaterialSiteChangeRelation(out, "Old Material Type", details.OldMaterialType, details.OldMaterialTypeID)
	renderMaterialSiteChangeRelation(out, "New Material Type", details.NewMaterialType, details.NewMaterialTypeID)
	renderMaterialSiteChangeRelation(out, "New Material Mix Design", details.NewMaterialMixDesign, details.NewMaterialMixDesignID)
	renderMaterialSiteChangeRelation(out, "Created By", details.CreatedBy, details.CreatedByID)

	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", details.UpdatedAt)
	}

	return nil
}

func renderMaterialSiteChangeRelation(out io.Writer, label, name, id string) {
	if name != "" {
		if id != "" {
			fmt.Fprintf(out, "%s: %s (%s)\n", label, name, id)
		} else {
			fmt.Fprintf(out, "%s: %s\n", label, name)
		}
		return
	}
	if id != "" {
		fmt.Fprintf(out, "%s ID: %s\n", label, id)
	}
}
