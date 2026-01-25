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

type equipmentMovementStopCompletionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementStopCompletionDetails struct {
	ID                string   `json:"id"`
	CompletedAt       string   `json:"completed_at,omitempty"`
	Latitude          string   `json:"latitude,omitempty"`
	Longitude         string   `json:"longitude,omitempty"`
	Note              string   `json:"note,omitempty"`
	StopID            string   `json:"stop_id,omitempty"`
	CreatedByID       string   `json:"created_by_id,omitempty"`
	FileAttachmentIDs []string `json:"file_attachment_ids,omitempty"`
}

func newEquipmentMovementStopCompletionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement stop completion details",
		Long: `Show the full details of a specific equipment movement stop completion.

Output Fields:
  ID              Stop completion identifier
  Stop            Stop ID
  Completed At    Completion timestamp
  Latitude        Completion latitude
  Longitude       Completion longitude
  Note            Completion note
  Created By      User who created the completion
  File Attachments  File attachment IDs

Arguments:
  <id>    The stop completion ID (required). You can find IDs using the list command.`,
		Example: `  # Show a stop completion
  xbe view equipment-movement-stop-completions show 123

  # Get JSON output
  xbe view equipment-movement-stop-completions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementStopCompletionsShow,
	}
	initEquipmentMovementStopCompletionsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementStopCompletionsCmd.AddCommand(newEquipmentMovementStopCompletionsShowCmd())
}

func initEquipmentMovementStopCompletionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementStopCompletionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseEquipmentMovementStopCompletionsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement stop completion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-movement-stop-completions]", "completed-at,latitude,longitude,note,stop,created-by,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-stop-completions/"+id, query)
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

	details := buildEquipmentMovementStopCompletionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementStopCompletionDetails(cmd, details)
}

func parseEquipmentMovementStopCompletionsShowOptions(cmd *cobra.Command) (equipmentMovementStopCompletionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementStopCompletionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementStopCompletionDetails(resp jsonAPISingleResponse) equipmentMovementStopCompletionDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := equipmentMovementStopCompletionDetails{
		ID:          resource.ID,
		CompletedAt: formatDateTime(stringAttr(attrs, "completed-at")),
		Latitude:    stringAttr(attrs, "latitude"),
		Longitude:   stringAttr(attrs, "longitude"),
		Note:        stringAttr(attrs, "note"),
	}

	if rel, ok := resource.Relationships["stop"]; ok && rel.Data != nil {
		details.StopID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				details.FileAttachmentIDs = append(details.FileAttachmentIDs, ref.ID)
			}
		}
	}

	return details
}

func renderEquipmentMovementStopCompletionDetails(cmd *cobra.Command, details equipmentMovementStopCompletionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.StopID != "" {
		fmt.Fprintf(out, "Stop: %s\n", details.StopID)
	}
	if details.CompletedAt != "" {
		fmt.Fprintf(out, "Completed At: %s\n", details.CompletedAt)
	}
	if details.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.Latitude)
	}
	if details.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.Longitude)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachments: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
