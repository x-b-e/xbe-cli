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

type objectiveStakeholderClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type objectiveStakeholderClassificationDetails struct {
	ID                          string   `json:"id"`
	ObjectiveID                 string   `json:"objective_id,omitempty"`
	StakeholderClassificationID string   `json:"stakeholder_classification_id,omitempty"`
	InterestDegree              *float64 `json:"interest_degree,omitempty"`
	QuoteIDs                    []string `json:"quote_ids,omitempty"`
}

func newObjectiveStakeholderClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show objective stakeholder classification details",
		Long: `Show the full details of an objective stakeholder classification.

Output Fields:
  ID
  Objective ID
  Stakeholder Classification ID
  Interest Degree
  Quote IDs

Arguments:
  <id>    The objective stakeholder classification ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an objective stakeholder classification
  xbe view objective-stakeholder-classifications show 123

  # Get JSON output
  xbe view objective-stakeholder-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runObjectiveStakeholderClassificationsShow,
	}
	initObjectiveStakeholderClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	objectiveStakeholderClassificationsCmd.AddCommand(newObjectiveStakeholderClassificationsShowCmd())
}

func initObjectiveStakeholderClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectiveStakeholderClassificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseObjectiveStakeholderClassificationsShowOptions(cmd)
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
		return fmt.Errorf("objective stakeholder classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[objective-stakeholder-classifications]", "interest-degree,objective,stakeholder-classification,quotes")

	body, _, err := client.Get(cmd.Context(), "/v1/objective-stakeholder-classifications/"+id, query)
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

	details := buildObjectiveStakeholderClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderObjectiveStakeholderClassificationDetails(cmd, details)
}

func parseObjectiveStakeholderClassificationsShowOptions(cmd *cobra.Command) (objectiveStakeholderClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectiveStakeholderClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildObjectiveStakeholderClassificationDetails(resp jsonAPISingleResponse) objectiveStakeholderClassificationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return objectiveStakeholderClassificationDetails{
		ID:                          resource.ID,
		ObjectiveID:                 relationshipIDFromMap(resource.Relationships, "objective"),
		StakeholderClassificationID: relationshipIDFromMap(resource.Relationships, "stakeholder-classification"),
		InterestDegree:              floatAttrPointer(attrs, "interest-degree"),
		QuoteIDs:                    relationshipIDsFromMap(resource.Relationships, "quotes"),
	}
}

func renderObjectiveStakeholderClassificationDetails(cmd *cobra.Command, details objectiveStakeholderClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ObjectiveID != "" {
		fmt.Fprintf(out, "Objective ID: %s\n", details.ObjectiveID)
	}
	if details.StakeholderClassificationID != "" {
		fmt.Fprintf(out, "Stakeholder Classification ID: %s\n", details.StakeholderClassificationID)
	}
	if details.InterestDegree != nil {
		fmt.Fprintf(out, "Interest Degree: %s\n", formatInterestDegree(details.InterestDegree))
	}
	if len(details.QuoteIDs) > 0 {
		fmt.Fprintf(out, "Quote IDs: %s\n", strings.Join(details.QuoteIDs, ", "))
	}

	return nil
}
