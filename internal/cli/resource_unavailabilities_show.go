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

type resourceUnavailabilitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type resourceUnavailabilityDetails struct {
	ID                string   `json:"id"`
	ResourceType      string   `json:"resource_type,omitempty"`
	ResourceID        string   `json:"resource_id,omitempty"`
	StartAt           string   `json:"start_at,omitempty"`
	EndAt             string   `json:"end_at,omitempty"`
	Description       string   `json:"description,omitempty"`
	CreatedByID       string   `json:"created_by_id,omitempty"`
	FileAttachmentIDs []string `json:"file_attachment_ids,omitempty"`
}

func newResourceUnavailabilitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show resource unavailability details",
		Long: `Show the full details of a resource unavailability.

Output Fields:
  ID
  Resource (type/id)
  Start At
  End At
  Description
  Created By (user ID)
  File Attachment IDs

Arguments:
  <id>    The resource unavailability ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a resource unavailability
  xbe view resource-unavailabilities show 123

  # Output as JSON
  xbe view resource-unavailabilities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runResourceUnavailabilitiesShow,
	}
	initResourceUnavailabilitiesShowFlags(cmd)
	return cmd
}

func init() {
	resourceUnavailabilitiesCmd.AddCommand(newResourceUnavailabilitiesShowCmd())
}

func initResourceUnavailabilitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runResourceUnavailabilitiesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseResourceUnavailabilitiesShowOptions(cmd)
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
		return fmt.Errorf("resource unavailability id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[resource-unavailabilities]", "start-at,end-at,description,resource,created-by,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/resource-unavailabilities/"+id, query)
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

	details := buildResourceUnavailabilityDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderResourceUnavailabilityDetails(cmd, details)
}

func parseResourceUnavailabilitiesShowOptions(cmd *cobra.Command) (resourceUnavailabilitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return resourceUnavailabilitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildResourceUnavailabilityDetails(resp jsonAPISingleResponse) resourceUnavailabilityDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := resourceUnavailabilityDetails{
		ID:          resource.ID,
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		details.ResourceType = rel.Data.Type
		details.ResourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDList(rel)
	}

	return details
}

func renderResourceUnavailabilityDetails(cmd *cobra.Command, details resourceUnavailabilityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ResourceType != "" || details.ResourceID != "" {
		resource := details.ResourceType
		if details.ResourceID != "" {
			if resource != "" {
				resource += "/"
			}
			resource += details.ResourceID
		}
		fmt.Fprintf(out, "Resource: %s\n", resource)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachments: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
