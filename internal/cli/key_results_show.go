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

type keyResultsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type keyResultDetails struct {
	ID                                   string    `json:"id"`
	Title                                string    `json:"title,omitempty"`
	TitleSummary                         string    `json:"title_summary,omitempty"`
	TitleSummaryExplicit                 string    `json:"title_summary_explicit,omitempty"`
	TitleSummaryImplicit                 string    `json:"title_summary_implicit,omitempty"`
	Status                               string    `json:"status,omitempty"`
	StartOn                              string    `json:"start_on,omitempty"`
	EndOn                                string    `json:"end_on,omitempty"`
	CompletionPercentage                 any       `json:"completion_percentage,omitempty"`
	CompletionPercentageCalculated       any       `json:"completion_percentage_calculated,omitempty"`
	ObjectiveID                          string    `json:"objective_id,omitempty"`
	ObjectiveName                        string    `json:"objective_name,omitempty"`
	OwnerID                              string    `json:"owner_id,omitempty"`
	OwnerName                            string    `json:"owner_name,omitempty"`
	CustomerSuccessResponsiblePersonID   string    `json:"customer_success_responsible_person_id,omitempty"`
	CustomerSuccessResponsiblePersonName string    `json:"customer_success_responsible_person_name,omitempty"`
	ChildObjectiveIDs                    []string  `json:"child_objective_ids,omitempty"`
	ActionItemKeyResultIDs               []string  `json:"action_item_key_result_ids,omitempty"`
	ActionItemIDs                        []string  `json:"action_item_ids,omitempty"`
	Comments                             []comment `json:"comments,omitempty"`
}

func newKeyResultsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show key result details",
		Long: `Show the full details of a key result.

Output Sections:
  Core fields (title, status, dates)
  Completion metrics
  Relationships (objective, owner, related items)
  Comments

Arguments:
  <id>    The key result ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a key result
  xbe view key-results show 123

  # Output as JSON
  xbe view key-results show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runKeyResultsShow,
	}
	initKeyResultsShowFlags(cmd)
	return cmd
}

func init() {
	keyResultsCmd.AddCommand(newKeyResultsShowCmd())
}

func initKeyResultsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseKeyResultsShowOptions(cmd)
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
		return fmt.Errorf("key result id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "objective,owner,customer-success-responsible-person,comments,comments.created-by")
	query.Set("fields[objectives]", "name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/key-results/"+id, query)
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

	details := buildKeyResultDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderKeyResultDetails(cmd, details)
}

func parseKeyResultsShowOptions(cmd *cobra.Command) (keyResultsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildKeyResultDetails(resp jsonAPISingleResponse) keyResultDetails {
	included := indexIncludedResources(resp.Included)
	attrs := resp.Data.Attributes

	details := keyResultDetails{
		ID:                             resp.Data.ID,
		Title:                          strings.TrimSpace(stringAttr(attrs, "title")),
		TitleSummary:                   stringAttr(attrs, "title-summary"),
		TitleSummaryExplicit:           stringAttr(attrs, "title-summary-explicit"),
		TitleSummaryImplicit:           stringAttr(attrs, "title-summary-implicit"),
		Status:                         stringAttr(attrs, "status"),
		StartOn:                        formatDate(stringAttr(attrs, "start-on")),
		EndOn:                          formatDate(stringAttr(attrs, "end-on")),
		CompletionPercentage:           attrs["completion-percentage"],
		CompletionPercentageCalculated: attrs["completion-percentage-calculated"],
	}

	if rel, ok := resp.Data.Relationships["objective"]; ok && rel.Data != nil {
		details.ObjectiveID = rel.Data.ID
		if obj, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ObjectiveName = strings.TrimSpace(stringAttr(obj.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["owner"]; ok && rel.Data != nil {
		details.OwnerID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OwnerName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["customer-success-responsible-person"]; ok && rel.Data != nil {
		details.CustomerSuccessResponsiblePersonID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerSuccessResponsiblePersonName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["children"]; ok && rel.raw != nil {
		details.ChildObjectiveIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["action-item-key-results"]; ok && rel.raw != nil {
		details.ActionItemKeyResultIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["action-items"]; ok && rel.raw != nil {
		details.ActionItemIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["comments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if c, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					cmt := comment{
						ID:        c.ID,
						Body:      strings.TrimSpace(stringAttr(c.Attributes, "body")),
						CreatedAt: formatDateTime(stringAttr(c.Attributes, "created-at")),
					}
					if userRel, ok := c.Relationships["created-by"]; ok && userRel.Data != nil {
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							cmt.CreatedByID = userRel.Data.ID
							cmt.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Comments = append(details.Comments, cmt)
				}
			}
		}
	}

	return details
}

func renderKeyResultDetails(cmd *cobra.Command, d keyResultDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Title != "" {
		fmt.Fprintf(out, "Title: %s\n", d.Title)
	}
	if d.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", d.Status)
	}
	if d.TitleSummary != "" {
		fmt.Fprintf(out, "Title Summary: %s\n", d.TitleSummary)
	}
	if d.TitleSummaryExplicit != "" {
		fmt.Fprintf(out, "Title Summary Explicit: %s\n", d.TitleSummaryExplicit)
	}
	if d.TitleSummaryImplicit != "" {
		fmt.Fprintf(out, "Title Summary Implicit: %s\n", d.TitleSummaryImplicit)
	}
	if d.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", d.StartOn)
	}
	if d.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", d.EndOn)
	}
	if d.CompletionPercentage != nil {
		fmt.Fprintf(out, "Completion Percentage: %s\n", formatAnyValue(d.CompletionPercentage))
	}
	if d.CompletionPercentageCalculated != nil {
		fmt.Fprintf(out, "Completion (Calculated): %s\n", formatAnyValue(d.CompletionPercentageCalculated))
	}

	if d.ObjectiveID != "" || d.OwnerID != "" || d.CustomerSuccessResponsiblePersonID != "" ||
		len(d.ChildObjectiveIDs) > 0 || len(d.ActionItemKeyResultIDs) > 0 || len(d.ActionItemIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Relationships:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ObjectiveName != "" {
			fmt.Fprintf(out, "  Objective: %s (ID: %s)\n", d.ObjectiveName, d.ObjectiveID)
		} else if d.ObjectiveID != "" {
			fmt.Fprintf(out, "  Objective ID: %s\n", d.ObjectiveID)
		}
		if d.OwnerName != "" {
			fmt.Fprintf(out, "  Owner: %s (ID: %s)\n", d.OwnerName, d.OwnerID)
		} else if d.OwnerID != "" {
			fmt.Fprintf(out, "  Owner ID: %s\n", d.OwnerID)
		}
		if d.CustomerSuccessResponsiblePersonName != "" {
			fmt.Fprintf(out, "  Customer Success Responsible: %s (ID: %s)\n", d.CustomerSuccessResponsiblePersonName, d.CustomerSuccessResponsiblePersonID)
		} else if d.CustomerSuccessResponsiblePersonID != "" {
			fmt.Fprintf(out, "  Customer Success Responsible ID: %s\n", d.CustomerSuccessResponsiblePersonID)
		}
		if len(d.ChildObjectiveIDs) > 0 {
			fmt.Fprintf(out, "  Child Objectives: %s\n", strings.Join(d.ChildObjectiveIDs, ", "))
		}
		if len(d.ActionItemKeyResultIDs) > 0 {
			fmt.Fprintf(out, "  Action Item Key Results: %s\n", strings.Join(d.ActionItemKeyResultIDs, ", "))
		}
		if len(d.ActionItemIDs) > 0 {
			fmt.Fprintf(out, "  Action Items: %s\n", strings.Join(d.ActionItemIDs, ", "))
		}
	}

	if len(d.Comments) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Comments (%d):\n", len(d.Comments))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, cmt := range d.Comments {
			author := cmt.CreatedBy
			if author == "" {
				author = cmt.CreatedByID
			}
			if author != "" {
				fmt.Fprintf(out, "  [%s] %s: %s\n", cmt.CreatedAt, author, cmt.Body)
			} else {
				fmt.Fprintf(out, "  [%s] %s\n", cmt.CreatedAt, cmt.Body)
			}
		}
	}

	return nil
}
