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

type openDoorIssuesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type openDoorIssueDetails struct {
	ID               string                    `json:"id"`
	Status           string                    `json:"status,omitempty"`
	Description      string                    `json:"description,omitempty"`
	CreatedAt        string                    `json:"created_at,omitempty"`
	UpdatedAt        string                    `json:"updated_at,omitempty"`
	OrganizationID   string                    `json:"organization_id,omitempty"`
	OrganizationType string                    `json:"organization_type,omitempty"`
	OrganizationName string                    `json:"organization_name,omitempty"`
	ReportedByID     string                    `json:"reported_by_id,omitempty"`
	ReportedByName   string                    `json:"reported_by_name,omitempty"`
	Comments         []openDoorIssueComment    `json:"comments,omitempty"`
	Attachments      []openDoorIssueAttachment `json:"attachments,omitempty"`
}

type openDoorIssueComment struct {
	ID          string `json:"id"`
	Body        string `json:"body,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
}

type openDoorIssueAttachment struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
}

func newOpenDoorIssuesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show open door issue details",
		Long: `Show the full details of a specific open door issue.

Output Fields:
  ID            Open door issue ID
  Status        Current status
  Description   Issue description
  Organization  Organization name (plus type/id)
  Reported By   Reporting user name (plus ID)
  Created       Created timestamp
  Updated       Updated timestamp
  Comments      Related comments
  Attachments   Related file attachments

Arguments:
  <id>    The open door issue ID (required). You can find IDs using the list command.`,
		Example: `  # Show an open door issue
  xbe view open-door-issues show 123

  # Get JSON output
  xbe view open-door-issues show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOpenDoorIssuesShow,
	}
	initOpenDoorIssuesShowFlags(cmd)
	return cmd
}

func init() {
	openDoorIssuesCmd.AddCommand(newOpenDoorIssuesShowCmd())
}

func initOpenDoorIssuesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenDoorIssuesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOpenDoorIssuesShowOptions(cmd)
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
		return fmt.Errorf("open door issue id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[open-door-issues]", "description,status,created-at,updated-at,organization,reported-by,comments,file-attachments")
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[comments]", "body,created-at,created-by")
	query.Set("fields[file-attachments]", "file-name,created-by")
	query.Set("include", "organization,reported-by,comments,comments.created-by,file-attachments,file-attachments.created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/open-door-issues/"+id, query)
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

	details := buildOpenDoorIssueDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOpenDoorIssueDetails(cmd, details)
}

func parseOpenDoorIssuesShowOptions(cmd *cobra.Command) (openDoorIssuesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openDoorIssuesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOpenDoorIssueDetails(resp jsonAPISingleResponse) openDoorIssueDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := openDoorIssueDetails{
		ID:          resp.Data.ID,
		Status:      strings.TrimSpace(stringAttr(attrs, "status")),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationID = rel.Data.ID
		details.OrganizationType = rel.Data.Type
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OrganizationName = strings.TrimSpace(stringAttr(org.Attributes, "company-name"))
			if details.OrganizationName == "" {
				details.OrganizationName = strings.TrimSpace(stringAttr(org.Attributes, "name"))
			}
		}
	}

	if rel, ok := resp.Data.Relationships["reported-by"]; ok && rel.Data != nil {
		details.ReportedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ReportedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["comments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if comment, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					entry := openDoorIssueComment{
						ID:        comment.ID,
						Body:      strings.TrimSpace(stringAttr(comment.Attributes, "body")),
						CreatedAt: formatDateTime(stringAttr(comment.Attributes, "created-at")),
					}
					if userRel, ok := comment.Relationships["created-by"]; ok && userRel.Data != nil {
						entry.CreatedByID = userRel.Data.ID
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							entry.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Comments = append(details.Comments, entry)
				}
			}
		}
	}

	if rel, ok := resp.Data.Relationships["file-attachments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if attachment, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					entry := openDoorIssueAttachment{
						ID:       attachment.ID,
						FileName: strings.TrimSpace(stringAttr(attachment.Attributes, "file-name")),
					}
					if userRel, ok := attachment.Relationships["created-by"]; ok && userRel.Data != nil {
						entry.CreatedByID = userRel.Data.ID
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							entry.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Attachments = append(details.Attachments, entry)
				}
			}
		}
	}

	return details
}

func renderOpenDoorIssueDetails(cmd *cobra.Command, details openDoorIssueDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}

	orgRef := formatResourceRef(details.OrganizationType, details.OrganizationID)
	if details.OrganizationName != "" {
		fmt.Fprintf(out, "Organization: %s\n", details.OrganizationName)
		if orgRef != "" {
			fmt.Fprintf(out, "Organization Ref: %s\n", orgRef)
		}
	} else if orgRef != "" {
		fmt.Fprintf(out, "Organization: %s\n", orgRef)
	}

	reportedBy := details.ReportedByName
	if reportedBy == "" {
		reportedBy = details.ReportedByID
	}
	if reportedBy != "" {
		fmt.Fprintf(out, "Reported By: %s\n", reportedBy)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", details.UpdatedAt)
	}

	if len(details.Comments) > 0 {
		fmt.Fprintf(out, "Comments (%d):\n", len(details.Comments))
		for _, comment := range details.Comments {
			author := comment.CreatedBy
			if author == "" {
				author = comment.CreatedByID
			}
			label := strings.TrimSpace(strings.Join([]string{author, comment.CreatedAt}, " "))
			if label != "" {
				fmt.Fprintf(out, "  - %s: %s\n", label, comment.Body)
			} else {
				fmt.Fprintf(out, "  - %s\n", comment.Body)
			}
		}
	}

	if len(details.Attachments) > 0 {
		fmt.Fprintf(out, "Attachments (%d):\n", len(details.Attachments))
		for _, attachment := range details.Attachments {
			name := attachment.FileName
			if name == "" {
				name = attachment.ID
			}
			owner := attachment.CreatedBy
			if owner == "" {
				owner = attachment.CreatedByID
			}
			if owner != "" {
				fmt.Fprintf(out, "  - %s (by %s)\n", name, owner)
			} else {
				fmt.Fprintf(out, "  - %s\n", name)
			}
		}
	}

	return nil
}
