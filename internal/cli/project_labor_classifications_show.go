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

type projectLaborClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectLaborClassificationDetails struct {
	ID                              string   `json:"id"`
	ProjectID                       string   `json:"project_id,omitempty"`
	ProjectName                     string   `json:"project_name,omitempty"`
	ProjectNumber                   string   `json:"project_number,omitempty"`
	LaborClassificationID           string   `json:"labor_classification_id,omitempty"`
	LaborClassificationName         string   `json:"labor_classification_name,omitempty"`
	LaborClassificationAbbreviation string   `json:"labor_classification_abbreviation,omitempty"`
	BasicHourlyRate                 string   `json:"basic_hourly_rate,omitempty"`
	FringeHourlyRate                string   `json:"fringe_hourly_rate,omitempty"`
	PrevailingHourlyRate            string   `json:"prevailing_hourly_rate,omitempty"`
	ProjectTrailerClassificationIDs []string `json:"project_trailer_classification_ids,omitempty"`
}

func newProjectLaborClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project labor classification details",
		Long: `Show the full details of a project labor classification.

Output Fields:
  ID                              Project labor classification identifier
  Project                         Project name or number
  Project ID                      Project identifier
  Labor Classification            Labor classification name
  Labor Classification ID         Labor classification identifier
  Basic Hourly Rate               Basic hourly rate
  Fringe Hourly Rate              Fringe hourly rate
  Prevailing Hourly Rate          Prevailing hourly rate
  Project Trailer Classification IDs  Related project trailer classification IDs

Arguments:
  <id>    Project labor classification ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project labor classification
  xbe view project-labor-classifications show 123

  # JSON output
  xbe view project-labor-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectLaborClassificationsShow,
	}
	initProjectLaborClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	projectLaborClassificationsCmd.AddCommand(newProjectLaborClassificationsShowCmd())
}

func initProjectLaborClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectLaborClassificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectLaborClassificationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("project labor classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-labor-classifications]", "project,labor-classification,project-trailer-classifications,basic-hourly-rate,fringe-hourly-rate,prevailing-hourly-rate")
	query.Set("include", "project,labor-classification")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[labor-classifications]", "name,abbreviation")

	body, _, err := client.Get(cmd.Context(), "/v1/project-labor-classifications/"+id, query)
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

	details := buildProjectLaborClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectLaborClassificationDetails(cmd, details)
}

func parseProjectLaborClassificationsShowOptions(cmd *cobra.Command) (projectLaborClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectLaborClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectLaborClassificationDetails(resp jsonAPISingleResponse) projectLaborClassificationDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := projectLaborClassificationDetails{
		ID:                   resp.Data.ID,
		BasicHourlyRate:      stringAttr(attrs, "basic-hourly-rate"),
		FringeHourlyRate:     stringAttr(attrs, "fringe-hourly-rate"),
		PrevailingHourlyRate: stringAttr(attrs, "prevailing-hourly-rate"),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(project.Attributes, "name")
			details.ProjectNumber = stringAttr(project.Attributes, "number")
		}
	}

	if rel, ok := resp.Data.Relationships["labor-classification"]; ok && rel.Data != nil {
		details.LaborClassificationID = rel.Data.ID
		if labor, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.LaborClassificationName = stringAttr(labor.Attributes, "name")
			details.LaborClassificationAbbreviation = stringAttr(labor.Attributes, "abbreviation")
		}
	}

	if rel, ok := resp.Data.Relationships["project-trailer-classifications"]; ok {
		details.ProjectTrailerClassificationIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectLaborClassificationDetails(cmd *cobra.Command, details projectLaborClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	projectLabel := details.ProjectName
	if projectLabel != "" && details.ProjectNumber != "" {
		projectLabel = fmt.Sprintf("%s (%s)", projectLabel, details.ProjectNumber)
	} else if projectLabel == "" {
		projectLabel = details.ProjectNumber
	}
	if projectLabel != "" {
		fmt.Fprintf(out, "Project: %s\n", projectLabel)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}

	laborLabel := details.LaborClassificationName
	if laborLabel != "" && details.LaborClassificationAbbreviation != "" {
		laborLabel = fmt.Sprintf("%s (%s)", laborLabel, details.LaborClassificationAbbreviation)
	} else if laborLabel == "" {
		laborLabel = details.LaborClassificationAbbreviation
	}
	if laborLabel != "" {
		fmt.Fprintf(out, "Labor Classification: %s\n", laborLabel)
	}
	if details.LaborClassificationID != "" {
		fmt.Fprintf(out, "Labor Classification ID: %s\n", details.LaborClassificationID)
	}

	fmt.Fprintf(out, "Basic Hourly Rate: %s\n", formatOptional(details.BasicHourlyRate))
	fmt.Fprintf(out, "Fringe Hourly Rate: %s\n", formatOptional(details.FringeHourlyRate))
	fmt.Fprintf(out, "Prevailing Hourly Rate: %s\n", formatOptional(details.PrevailingHourlyRate))

	if len(details.ProjectTrailerClassificationIDs) > 0 {
		fmt.Fprintf(out, "Project Trailer Classification IDs: %s\n", strings.Join(details.ProjectTrailerClassificationIDs, ", "))
	} else {
		fmt.Fprintln(out, "Project Trailer Classification IDs: (none)")
	}

	return nil
}
