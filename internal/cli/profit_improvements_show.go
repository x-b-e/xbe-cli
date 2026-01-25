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

type profitImprovementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type profitImprovementDetails struct {
	ID                          string           `json:"id"`
	Title                       string           `json:"title,omitempty"`
	Description                 string           `json:"description,omitempty"`
	Status                      string           `json:"status,omitempty"`
	AmountEstimated             any              `json:"amount_estimated,omitempty"`
	ImpactFrequencyEstimated    string           `json:"impact_frequency_estimated,omitempty"`
	ImpactIntervalEstimated     string           `json:"impact_interval_estimated,omitempty"`
	ImpactStartOnEstimated      string           `json:"impact_start_on_estimated,omitempty"`
	ImpactEndOnEstimated        string           `json:"impact_end_on_estimated,omitempty"`
	AmountValidated             any              `json:"amount_validated,omitempty"`
	ImpactFrequencyValidated    string           `json:"impact_frequency_validated,omitempty"`
	ImpactIntervalValidated     string           `json:"impact_interval_validated,omitempty"`
	ImpactStartOnValidated      string           `json:"impact_start_on_validated,omitempty"`
	ImpactEndOnValidated        string           `json:"impact_end_on_validated,omitempty"`
	GainShareFeePercentage      any              `json:"gain_share_fee_percentage,omitempty"`
	GainShareFeeStartOn         string           `json:"gain_share_fee_start_on,omitempty"`
	GainShareFeeEndOn           string           `json:"gain_share_fee_end_on,omitempty"`
	CreatedAt                   string           `json:"created_at,omitempty"`
	UpdatedAt                   string           `json:"updated_at,omitempty"`
	ProfitImprovementCategoryID string           `json:"profit_improvement_category_id,omitempty"`
	ProfitImprovementCategory   string           `json:"profit_improvement_category,omitempty"`
	OriginalID                  string           `json:"original_id,omitempty"`
	CreatedByID                 string           `json:"created_by_id,omitempty"`
	CreatedBy                   string           `json:"created_by,omitempty"`
	OwnedByID                   string           `json:"owned_by_id,omitempty"`
	OwnedBy                     string           `json:"owned_by,omitempty"`
	ValidatedByID               string           `json:"validated_by_id,omitempty"`
	ValidatedBy                 string           `json:"validated_by,omitempty"`
	OrganizationType            string           `json:"organization_type,omitempty"`
	OrganizationID              string           `json:"organization_id,omitempty"`
	Organization                string           `json:"organization,omitempty"`
	BrokerID                    string           `json:"broker_id,omitempty"`
	Broker                      string           `json:"broker,omitempty"`
	DuplicatedIDs               []string         `json:"duplicated_ids,omitempty"`
	SubscriptionIDs             []string         `json:"profit_improvement_subscription_ids,omitempty"`
	Comments                    []comment        `json:"comments,omitempty"`
	Attachments                 []fileAttachment `json:"attachments,omitempty"`
}

func newProfitImprovementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show profit improvement details",
		Long: `Show the full details of a profit improvement.

Output Sections:
  Core fields (title, status, description)
  Impact estimates and validation
  Gain share details
  Ownership and relationships
  Comments and attachments

Arguments:
  <id>    The profit improvement ID (required). Use the list command to find IDs.`,
		Example: `  # Show a profit improvement
  xbe view profit-improvements show 123

  # Show as JSON
  xbe view profit-improvements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProfitImprovementsShow,
	}
	initProfitImprovementsShowFlags(cmd)
	return cmd
}

func init() {
	profitImprovementsCmd.AddCommand(newProfitImprovementsShowCmd())
}

func initProfitImprovementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProfitImprovementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProfitImprovementsShowOptions(cmd)
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
		return fmt.Errorf("profit improvement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "profit-improvement-category,created-by,owned-by,validated-by,organization,broker,comments,comments.created-by,file-attachments,file-attachments.created-by")
	query.Set("fields[users]", "name")
	query.Set("fields[profit-improvement-categories]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/profit-improvements/"+id, query)
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

	details := buildProfitImprovementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProfitImprovementDetails(cmd, details)
}

func parseProfitImprovementsShowOptions(cmd *cobra.Command) (profitImprovementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return profitImprovementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProfitImprovementDetails(resp jsonAPISingleResponse) profitImprovementDetails {
	included := indexIncludedResources(resp.Included)
	attrs := resp.Data.Attributes

	details := profitImprovementDetails{
		ID:                       resp.Data.ID,
		Title:                    stringAttr(attrs, "title"),
		Description:              stringAttr(attrs, "description"),
		Status:                   stringAttr(attrs, "status"),
		AmountEstimated:          attrs["amount-estimated"],
		ImpactFrequencyEstimated: stringAttr(attrs, "impact-frequency-estimated"),
		ImpactIntervalEstimated:  stringAttr(attrs, "impact-interval-estimated"),
		ImpactStartOnEstimated:   formatDate(stringAttr(attrs, "impact-start-on-estimated")),
		ImpactEndOnEstimated:     formatDate(stringAttr(attrs, "impact-end-on-estimated")),
		AmountValidated:          attrs["amount-validated"],
		ImpactFrequencyValidated: stringAttr(attrs, "impact-frequency-validated"),
		ImpactIntervalValidated:  stringAttr(attrs, "impact-interval-validated"),
		ImpactStartOnValidated:   formatDate(stringAttr(attrs, "impact-start-on-validated")),
		ImpactEndOnValidated:     formatDate(stringAttr(attrs, "impact-end-on-validated")),
		GainShareFeePercentage:   attrs["gain-share-fee-percentage"],
		GainShareFeeStartOn:      formatDate(stringAttr(attrs, "gain-share-fee-start-on")),
		GainShareFeeEndOn:        formatDate(stringAttr(attrs, "gain-share-fee-end-on")),
		CreatedAt:                formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["profit-improvement-category"]; ok && rel.Data != nil {
		details.ProfitImprovementCategoryID = rel.Data.ID
		if cat, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProfitImprovementCategory = stringAttr(cat.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["original"]; ok && rel.Data != nil {
		details.OriginalID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedBy = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["owned-by"]; ok && rel.Data != nil {
		details.OwnedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OwnedBy = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["validated-by"]; ok && rel.Data != nil {
		details.ValidatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ValidatedBy = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Organization = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
				stringAttr(org.Attributes, "title"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Broker = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resp.Data.Relationships["duplicated"]; ok && rel.raw != nil {
		details.DuplicatedIDs = extractRelationshipIDs(rel)
	}

	if rel, ok := resp.Data.Relationships["profit-improvement-subscriptions"]; ok && rel.raw != nil {
		details.SubscriptionIDs = extractRelationshipIDs(rel)
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
						cmt.CreatedByID = userRel.Data.ID
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							cmt.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Comments = append(details.Comments, cmt)
				}
			}
		}
	}

	if rel, ok := resp.Data.Relationships["file-attachments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if fa, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					att := fileAttachment{
						ID:          fa.ID,
						Filename:    strings.TrimSpace(stringAttr(fa.Attributes, "filename")),
						ContentType: stringAttr(fa.Attributes, "content-type"),
					}
					if userRel, ok := fa.Relationships["created-by"]; ok && userRel.Data != nil {
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							att.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Attachments = append(details.Attachments, att)
				}
			}
		}
	}

	return details
}

func extractRelationshipIDs(rel jsonAPIRelationship) []string {
	if rel.raw == nil {
		return nil
	}
	var refs []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &refs); err != nil {
		return nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.ID != "" {
			ids = append(ids, ref.ID)
		}
	}
	return ids
}

func renderProfitImprovementDetails(cmd *cobra.Command, d profitImprovementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Title != "" {
		fmt.Fprintf(out, "Title: %s\n", d.Title)
	}
	if d.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", d.Status)
	}
	if d.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", d.Description)
	}
	if d.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", d.CreatedAt)
	}
	if d.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", d.UpdatedAt)
	}

	if d.ProfitImprovementCategoryID != "" || d.OrganizationID != "" || d.OwnedByID != "" || d.CreatedByID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Relationships:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ProfitImprovementCategory != "" {
			fmt.Fprintf(out, "  Category: %s (ID: %s)\n", d.ProfitImprovementCategory, d.ProfitImprovementCategoryID)
		} else if d.ProfitImprovementCategoryID != "" {
			fmt.Fprintf(out, "  Category ID: %s\n", d.ProfitImprovementCategoryID)
		}
		if d.Organization != "" {
			fmt.Fprintf(out, "  Organization: %s (%s, ID: %s)\n", d.Organization, d.OrganizationType, d.OrganizationID)
		} else if d.OrganizationID != "" {
			fmt.Fprintf(out, "  Organization: %s/%s\n", d.OrganizationType, d.OrganizationID)
		}
		if d.Broker != "" {
			fmt.Fprintf(out, "  Broker: %s (ID: %s)\n", d.Broker, d.BrokerID)
		} else if d.BrokerID != "" {
			fmt.Fprintf(out, "  Broker ID: %s\n", d.BrokerID)
		}
		if d.CreatedBy != "" {
			fmt.Fprintf(out, "  Created By: %s (ID: %s)\n", d.CreatedBy, d.CreatedByID)
		} else if d.CreatedByID != "" {
			fmt.Fprintf(out, "  Created By ID: %s\n", d.CreatedByID)
		}
		if d.OwnedBy != "" {
			fmt.Fprintf(out, "  Owned By: %s (ID: %s)\n", d.OwnedBy, d.OwnedByID)
		} else if d.OwnedByID != "" {
			fmt.Fprintf(out, "  Owned By ID: %s\n", d.OwnedByID)
		}
		if d.ValidatedBy != "" {
			fmt.Fprintf(out, "  Validated By: %s (ID: %s)\n", d.ValidatedBy, d.ValidatedByID)
		} else if d.ValidatedByID != "" {
			fmt.Fprintf(out, "  Validated By ID: %s\n", d.ValidatedByID)
		}
		if d.OriginalID != "" {
			fmt.Fprintf(out, "  Original ID: %s\n", d.OriginalID)
		}
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Estimated Impact:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Amount: %s\n", formatAnyValue(d.AmountEstimated))
	if d.ImpactFrequencyEstimated != "" {
		fmt.Fprintf(out, "  Frequency: %s\n", d.ImpactFrequencyEstimated)
	}
	if d.ImpactIntervalEstimated != "" {
		fmt.Fprintf(out, "  Interval: %s\n", d.ImpactIntervalEstimated)
	}
	if d.ImpactStartOnEstimated != "" {
		fmt.Fprintf(out, "  Start: %s\n", d.ImpactStartOnEstimated)
	}
	if d.ImpactEndOnEstimated != "" {
		fmt.Fprintf(out, "  End: %s\n", d.ImpactEndOnEstimated)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Validated Impact:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Amount: %s\n", formatAnyValue(d.AmountValidated))
	if d.ImpactFrequencyValidated != "" {
		fmt.Fprintf(out, "  Frequency: %s\n", d.ImpactFrequencyValidated)
	}
	if d.ImpactIntervalValidated != "" {
		fmt.Fprintf(out, "  Interval: %s\n", d.ImpactIntervalValidated)
	}
	if d.ImpactStartOnValidated != "" {
		fmt.Fprintf(out, "  Start: %s\n", d.ImpactStartOnValidated)
	}
	if d.ImpactEndOnValidated != "" {
		fmt.Fprintf(out, "  End: %s\n", d.ImpactEndOnValidated)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Gain Share:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Percentage: %s\n", formatAnyValue(d.GainShareFeePercentage))
	if d.GainShareFeeStartOn != "" {
		fmt.Fprintf(out, "  Start: %s\n", d.GainShareFeeStartOn)
	}
	if d.GainShareFeeEndOn != "" {
		fmt.Fprintf(out, "  End: %s\n", d.GainShareFeeEndOn)
	}

	if len(d.DuplicatedIDs) > 0 || len(d.SubscriptionIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Related IDs:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if len(d.DuplicatedIDs) > 0 {
			fmt.Fprintf(out, "  Duplicates: %s\n", strings.Join(d.DuplicatedIDs, ", "))
		}
		if len(d.SubscriptionIDs) > 0 {
			fmt.Fprintf(out, "  Subscriptions: %s\n", strings.Join(d.SubscriptionIDs, ", "))
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

	if len(d.Attachments) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Attachments (%d):\n", len(d.Attachments))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, att := range d.Attachments {
			if att.CreatedBy != "" {
				fmt.Fprintf(out, "  %s (%s) by %s\n", att.Filename, att.ContentType, att.CreatedBy)
			} else {
				fmt.Fprintf(out, "  %s (%s)\n", att.Filename, att.ContentType)
			}
		}
	}

	return nil
}
