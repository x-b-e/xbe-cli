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

type developerTruckerCertificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type developerTruckerCertificationDetails struct {
	ID                              string   `json:"id"`
	StartOn                         string   `json:"start_on,omitempty"`
	EndOn                           string   `json:"end_on,omitempty"`
	DefaultMultiplier               string   `json:"default_multiplier,omitempty"`
	DeveloperID                     string   `json:"developer_id,omitempty"`
	DeveloperName                   string   `json:"developer_name,omitempty"`
	TruckerID                       string   `json:"trucker_id,omitempty"`
	TruckerName                     string   `json:"trucker_name,omitempty"`
	ClassificationID                string   `json:"classification_id,omitempty"`
	ClassificationName              string   `json:"classification_name,omitempty"`
	MultiplierIDs                   []string `json:"multiplier_ids,omitempty"`
	CurrentUserCanCreateMultipliers bool     `json:"current_user_can_create_multipliers"`
	CurrentUserCanDestroy           bool     `json:"current_user_can_destroy"`
	CanDelete                       bool     `json:"can_delete"`
}

func newDeveloperTruckerCertificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show developer trucker certification details",
		Long: `Show the full details of a developer trucker certification.

Output Fields:
  ID
  Start / End
  Default Multiplier
  Developer (name + ID)
  Trucker (name + ID)
  Classification (name + ID)
  Multipliers (IDs)
  Permissions (current user can create multipliers/destroy, can delete)

Arguments:
  <id>    The developer trucker certification ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a developer trucker certification
  xbe view developer-trucker-certifications show 123

  # Show as JSON
  xbe view developer-trucker-certifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeveloperTruckerCertificationsShow,
	}
	initDeveloperTruckerCertificationsShowFlags(cmd)
	return cmd
}

func init() {
	developerTruckerCertificationsCmd.AddCommand(newDeveloperTruckerCertificationsShowCmd())
}

func initDeveloperTruckerCertificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperTruckerCertificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDeveloperTruckerCertificationsShowOptions(cmd)
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
		return fmt.Errorf("developer trucker certification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[developer-trucker-certifications]", "start-on,end-on,default-multiplier,developer,trucker,classification,multipliers")
	query.Set("include", "developer,trucker,classification")
	query.Set("fields[developers]", "name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developer-trucker-certification-classifications]", "name")
	query.Set("meta[developer-trucker-certification]", "current_user_can_create_multipliers,current_user_can_destroy,can_delete")

	body, _, err := client.Get(cmd.Context(), "/v1/developer-trucker-certifications/"+id, query)
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

	details := buildDeveloperTruckerCertificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeveloperTruckerCertificationDetails(cmd, details)
}

func parseDeveloperTruckerCertificationsShowOptions(cmd *cobra.Command) (developerTruckerCertificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerTruckerCertificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeveloperTruckerCertificationDetails(resp jsonAPISingleResponse) developerTruckerCertificationDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	resource := resp.Data
	details := developerTruckerCertificationDetails{
		ID:                              resource.ID,
		StartOn:                         stringAttr(resource.Attributes, "start-on"),
		EndOn:                           stringAttr(resource.Attributes, "end-on"),
		DefaultMultiplier:               stringAttr(resource.Attributes, "default-multiplier"),
		CurrentUserCanCreateMultipliers: boolAttr(resource.Meta, "current_user_can_create_multipliers"),
		CurrentUserCanDestroy:           boolAttr(resource.Meta, "current_user_can_destroy"),
		CanDelete:                       boolAttr(resource.Meta, "can_delete"),
	}

	if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
		details.DeveloperID = rel.Data.ID
		if developer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DeveloperName = stringAttr(developer.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = firstNonEmpty(
				stringAttr(trucker.Attributes, "company-name"),
				stringAttr(trucker.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["classification"]; ok && rel.Data != nil {
		details.ClassificationID = rel.Data.ID
		if classification, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ClassificationName = stringAttr(classification.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["multipliers"]; ok {
		ids := relationshipIDs(rel)
		if len(ids) > 0 {
			details.MultiplierIDs = make([]string, 0, len(ids))
			for _, id := range ids {
				details.MultiplierIDs = append(details.MultiplierIDs, id.ID)
			}
		}
	}

	return details
}

func renderDeveloperTruckerCertificationDetails(cmd *cobra.Command, details developerTruckerCertificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", details.EndOn)
	}
	if details.DefaultMultiplier != "" {
		fmt.Fprintf(out, "Default Multiplier: %s\n", details.DefaultMultiplier)
	}
	if details.DeveloperName != "" {
		fmt.Fprintf(out, "Developer Name: %s\n", details.DeveloperName)
	}
	if details.DeveloperID != "" {
		fmt.Fprintf(out, "Developer ID: %s\n", details.DeveloperID)
	}
	if details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker Name: %s\n", details.TruckerName)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.ClassificationName != "" {
		fmt.Fprintf(out, "Classification Name: %s\n", details.ClassificationName)
	}
	if details.ClassificationID != "" {
		fmt.Fprintf(out, "Classification ID: %s\n", details.ClassificationID)
	}
	if len(details.MultiplierIDs) > 0 {
		fmt.Fprintf(out, "Multiplier IDs: %s\n", strings.Join(details.MultiplierIDs, ", "))
	}
	fmt.Fprintf(out, "Current User Can Create Multipliers: %t\n", details.CurrentUserCanCreateMultipliers)
	fmt.Fprintf(out, "Current User Can Destroy: %t\n", details.CurrentUserCanDestroy)
	fmt.Fprintf(out, "Can Delete: %t\n", details.CanDelete)

	return nil
}
