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

type resourceClassificationProjectCostClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type resourceClassificationProjectCostClassificationDetails struct {
	ID                         string `json:"id"`
	ResourceClassificationType string `json:"resource_classification_type,omitempty"`
	ResourceClassificationID   string `json:"resource_classification_id,omitempty"`
	ProjectCostClassification  string `json:"project_cost_classification_id,omitempty"`
	Broker                     string `json:"broker_id,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

func newResourceClassificationProjectCostClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show resource classification project cost classification details",
		Long: `Show the full details of a resource classification project cost classification.

Output Fields:
  ID
  Resource Classification Type
  Resource Classification ID
  Project Cost Classification ID
  Broker ID
  Created At
  Updated At

Arguments:
  <id>    The association ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a resource classification project cost classification
  xbe view resource-classification-project-cost-classifications show 123

  # JSON output
  xbe view resource-classification-project-cost-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runResourceClassificationProjectCostClassificationsShow,
	}
	initResourceClassificationProjectCostClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	resourceClassificationProjectCostClassificationsCmd.AddCommand(newResourceClassificationProjectCostClassificationsShowCmd())
}

func initResourceClassificationProjectCostClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runResourceClassificationProjectCostClassificationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseResourceClassificationProjectCostClassificationsShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("resource classification project cost classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[resource-classification-project-cost-classifications]", "resource-classification,project-cost-classification,broker,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/resource-classification-project-cost-classifications/"+id, query)
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

	details := buildResourceClassificationProjectCostClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderResourceClassificationProjectCostClassificationDetails(cmd, details)
}

func parseResourceClassificationProjectCostClassificationsShowOptions(cmd *cobra.Command) (resourceClassificationProjectCostClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return resourceClassificationProjectCostClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildResourceClassificationProjectCostClassificationDetails(resp jsonAPISingleResponse) resourceClassificationProjectCostClassificationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := resourceClassificationProjectCostClassificationDetails{
		ID:                        resource.ID,
		ProjectCostClassification: relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
		Broker:                    relationshipIDFromMap(resource.Relationships, "broker"),
		CreatedAt:                 formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                 formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["resource-classification"]; ok && rel.Data != nil {
		details.ResourceClassificationType = rel.Data.Type
		details.ResourceClassificationID = rel.Data.ID
	}

	return details
}

func renderResourceClassificationProjectCostClassificationDetails(cmd *cobra.Command, details resourceClassificationProjectCostClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ResourceClassificationType != "" {
		fmt.Fprintf(out, "Resource Classification Type: %s\n", details.ResourceClassificationType)
	}
	if details.ResourceClassificationID != "" {
		fmt.Fprintf(out, "Resource Classification ID: %s\n", details.ResourceClassificationID)
	}
	if details.ProjectCostClassification != "" {
		fmt.Fprintf(out, "Project Cost Classification ID: %s\n", details.ProjectCostClassification)
	}
	if details.Broker != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.Broker)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
