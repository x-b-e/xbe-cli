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

type meetingsShowOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	IncludeDescriptionLinks       bool
	IncludeSafetyMeetingPlansInfo bool
}

type meetingDetails struct {
	ID                        string   `json:"id"`
	Subject                   string   `json:"subject,omitempty"`
	Description               string   `json:"description,omitempty"`
	Transcript                string   `json:"transcript,omitempty"`
	Summary                   string   `json:"summary,omitempty"`
	StartAt                   string   `json:"start_at,omitempty"`
	EndAt                     string   `json:"end_at,omitempty"`
	ExplicitTimeZoneID        string   `json:"explicit_time_zone_id,omitempty"`
	TimeZoneID                string   `json:"time_zone_id,omitempty"`
	CurrentUserCanUpdate      bool     `json:"current_user_can_update"`
	Address                   string   `json:"address,omitempty"`
	IsAddressFormattedAddress bool     `json:"is_address_formatted_address"`
	AddressFormatted          string   `json:"address_formatted,omitempty"`
	AddressCity               string   `json:"address_city,omitempty"`
	AddressStateCode          string   `json:"address_state_code,omitempty"`
	AddressTimeZoneID         string   `json:"address_time_zone_id,omitempty"`
	AddressLatitude           string   `json:"address_latitude,omitempty"`
	AddressLongitude          string   `json:"address_longitude,omitempty"`
	AddressPlaceID            string   `json:"address_place_id,omitempty"`
	AddressPlusCode           string   `json:"address_plus_code,omitempty"`
	SkipAddressGeocoding      bool     `json:"skip_address_geocoding,omitempty"`
	OrganizationType          string   `json:"organization_type,omitempty"`
	OrganizationID            string   `json:"organization_id,omitempty"`
	OrganizationName          string   `json:"organization_name,omitempty"`
	OrganizerID               string   `json:"organizer_id,omitempty"`
	OrganizerName             string   `json:"organizer_name,omitempty"`
	CommentIDs                []string `json:"comment_ids,omitempty"`
	MeetingAttendeeIDs        []string `json:"meeting_attendee_ids,omitempty"`
	SafetyMeetingJobPlanIDs   []string `json:"safety_meeting_job_production_plan_ids,omitempty"`
	ActionItemIDs             []string `json:"action_item_ids,omitempty"`
	FileAttachmentIDs         []string `json:"file_attachment_ids,omitempty"`
	DescriptionLinks          any      `json:"description_links,omitempty"`
	SafetyMeetingJobPlansInfo any      `json:"safety_meeting_job_production_plans_info,omitempty"`
}

func newMeetingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show meeting details",
		Long: `Show the full details of a specific meeting.

Output Fields:
  Core fields, schedule, location, time zone, and permission metadata
  Organization and organizer context
  Related attendees, comments, action items, and attachments

Arguments:
  <id>  The meeting ID (required). Use the list command to find IDs.

Flags:
  --include-description-links        Include description links meta
  --include-safety-meeting-plans-info Include safety meeting job plan info meta

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show meeting details
  xbe view meetings show 123

  # Include description link metadata
  xbe view meetings show 123 --include-description-links

  # Include safety meeting job plan info
  xbe view meetings show 123 --include-safety-meeting-plans-info

  # JSON output
  xbe view meetings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMeetingsShow,
	}
	initMeetingsShowFlags(cmd)
	return cmd
}

func init() {
	meetingsCmd.AddCommand(newMeetingsShowCmd())
}

func initMeetingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("include-description-links", false, "Include description links meta")
	cmd.Flags().Bool("include-safety-meeting-plans-info", false, "Include safety meeting job plan info meta")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMeetingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMeetingsShowOptions(cmd)
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
		return fmt.Errorf("meeting id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	include := []string{
		"organization",
		"organizer",
		"meeting-attendees",
		"meeting-attendees.user",
		"comments",
		"comments.created-by",
		"action-items",
		"safety-meeting-job-production-plans",
		"file-attachments",
		"file-attachments.created-by",
	}
	query.Set("include", strings.Join(include, ","))
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	metaOptions := []string{}
	if opts.IncludeDescriptionLinks {
		metaOptions = append(metaOptions, "description-links")
	}
	if opts.IncludeSafetyMeetingPlansInfo {
		metaOptions = append(metaOptions, "safety-meeting-job-production-plans-info")
	}
	if len(metaOptions) > 0 {
		query.Set("meta[meeting]", strings.Join(metaOptions, ","))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/meetings/"+id, query)
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

	details := buildMeetingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMeetingDetails(cmd, details)
}

func parseMeetingsShowOptions(cmd *cobra.Command) (meetingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	includeDescriptionLinks, _ := cmd.Flags().GetBool("include-description-links")
	includeSafetyMeetingPlansInfo, _ := cmd.Flags().GetBool("include-safety-meeting-plans-info")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return meetingsShowOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		IncludeDescriptionLinks:       includeDescriptionLinks,
		IncludeSafetyMeetingPlansInfo: includeSafetyMeetingPlansInfo,
	}, nil
}

func buildMeetingDetails(resp jsonAPISingleResponse) meetingDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := meetingDetails{
		ID:                        resource.ID,
		Subject:                   strings.TrimSpace(stringAttr(attrs, "subject")),
		Description:               strings.TrimSpace(stringAttr(attrs, "description")),
		Transcript:                strings.TrimSpace(stringAttr(attrs, "transcript")),
		Summary:                   strings.TrimSpace(stringAttr(attrs, "summary")),
		StartAt:                   formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                     formatDateTime(stringAttr(attrs, "end-at")),
		ExplicitTimeZoneID:        strings.TrimSpace(stringAttr(attrs, "explicit-time-zone-id")),
		TimeZoneID:                strings.TrimSpace(stringAttr(attrs, "time-zone-id")),
		CurrentUserCanUpdate:      boolAttr(attrs, "current-user-can-update"),
		Address:                   strings.TrimSpace(stringAttr(attrs, "address")),
		IsAddressFormattedAddress: boolAttr(attrs, "is-address-formatted-address"),
		AddressFormatted:          strings.TrimSpace(stringAttr(attrs, "address-formatted")),
		AddressCity:               strings.TrimSpace(stringAttr(attrs, "address-city")),
		AddressStateCode:          strings.TrimSpace(stringAttr(attrs, "address-state-code")),
		AddressTimeZoneID:         strings.TrimSpace(stringAttr(attrs, "address-time-zone-id")),
		AddressLatitude:           strings.TrimSpace(stringAttr(attrs, "address-latitude")),
		AddressLongitude:          strings.TrimSpace(stringAttr(attrs, "address-longitude")),
		AddressPlaceID:            strings.TrimSpace(stringAttr(attrs, "address-place-id")),
		AddressPlusCode:           strings.TrimSpace(stringAttr(attrs, "address-plus-code")),
		SkipAddressGeocoding:      boolAttr(attrs, "skip-address-geocoding"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OrganizationName = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["organizer"]; ok && rel.Data != nil {
		details.OrganizerID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.OrganizerName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	details.CommentIDs = relationshipIDsFromMap(resource.Relationships, "comments")
	details.MeetingAttendeeIDs = relationshipIDsFromMap(resource.Relationships, "meeting-attendees")
	details.SafetyMeetingJobPlanIDs = relationshipIDsFromMap(resource.Relationships, "safety-meeting-job-production-plans")
	details.ActionItemIDs = relationshipIDsFromMap(resource.Relationships, "action-items")
	details.FileAttachmentIDs = relationshipIDsFromMap(resource.Relationships, "file-attachments")

	if resource.Meta != nil {
		if value, ok := resource.Meta["description_links"]; ok {
			details.DescriptionLinks = value
		} else if value, ok := resource.Meta["description-links"]; ok {
			details.DescriptionLinks = value
		}
		if value, ok := resource.Meta["safety_meeting_job_production_plans_info"]; ok {
			details.SafetyMeetingJobPlansInfo = value
		} else if value, ok := resource.Meta["safety-meeting-job-production-plans-info"]; ok {
			details.SafetyMeetingJobPlansInfo = value
		}
	}

	return details
}

func renderMeetingDetails(cmd *cobra.Command, details meetingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Subject != "" {
		fmt.Fprintf(out, "Subject: %s\n", details.Subject)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.ExplicitTimeZoneID != "" {
		fmt.Fprintf(out, "Explicit Time Zone: %s\n", details.ExplicitTimeZoneID)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", details.TimeZoneID)
	}
	fmt.Fprintf(out, "Current User Can Update: %s\n", formatYesNo(details.CurrentUserCanUpdate))

	if details.Address != "" {
		fmt.Fprintf(out, "Address: %s\n", details.Address)
	}
	if details.AddressFormatted != "" {
		fmt.Fprintf(out, "Address Formatted: %s\n", details.AddressFormatted)
	}
	if details.AddressCity != "" || details.AddressStateCode != "" {
		fmt.Fprintf(out, "Address City/State: %s %s\n", details.AddressCity, details.AddressStateCode)
	}
	if details.AddressTimeZoneID != "" {
		fmt.Fprintf(out, "Address Time Zone: %s\n", details.AddressTimeZoneID)
	}
	if details.AddressLatitude != "" || details.AddressLongitude != "" {
		fmt.Fprintf(out, "Address Coordinates: %s, %s\n", details.AddressLatitude, details.AddressLongitude)
	}
	if details.AddressPlaceID != "" {
		fmt.Fprintf(out, "Address Place ID: %s\n", details.AddressPlaceID)
	}
	if details.AddressPlusCode != "" {
		fmt.Fprintf(out, "Address Plus Code: %s\n", details.AddressPlusCode)
	}
	if details.IsAddressFormattedAddress {
		fmt.Fprintf(out, "Address Is Formatted: %s\n", formatYesNo(details.IsAddressFormattedAddress))
	}
	if details.SkipAddressGeocoding {
		fmt.Fprintf(out, "Skip Address Geocoding: %s\n", formatYesNo(details.SkipAddressGeocoding))
	}

	if details.OrganizationID != "" {
		orgLabel := details.OrganizationName
		if orgLabel == "" {
			orgLabel = fmt.Sprintf("%s:%s", details.OrganizationType, details.OrganizationID)
		}
		fmt.Fprintf(out, "Organization: %s\n", orgLabel)
	}
	if details.OrganizerID != "" {
		organizerLabel := details.OrganizerName
		if organizerLabel == "" {
			organizerLabel = details.OrganizerID
		}
		fmt.Fprintf(out, "Organizer: %s\n", organizerLabel)
	}
	if details.Transcript != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Transcript:")
		fmt.Fprintln(out, details.Transcript)
	}
	if details.Summary != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Summary:")
		fmt.Fprintln(out, details.Summary)
	}
	if len(details.MeetingAttendeeIDs) > 0 {
		fmt.Fprintf(out, "Meeting Attendee IDs: %s\n", strings.Join(details.MeetingAttendeeIDs, ", "))
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}
	if len(details.ActionItemIDs) > 0 {
		fmt.Fprintf(out, "Action Item IDs: %s\n", strings.Join(details.ActionItemIDs, ", "))
	}
	if len(details.SafetyMeetingJobPlanIDs) > 0 {
		fmt.Fprintf(out, "Safety Meeting Job Plan IDs: %s\n", strings.Join(details.SafetyMeetingJobPlanIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}
	if details.DescriptionLinks != nil {
		payload, err := json.MarshalIndent(details.DescriptionLinks, "", "  ")
		if err == nil {
			fmt.Fprintln(out, "Description Links:")
			fmt.Fprintln(out, string(payload))
		}
	}
	if details.SafetyMeetingJobPlansInfo != nil {
		payload, err := json.MarshalIndent(details.SafetyMeetingJobPlansInfo, "", "  ")
		if err == nil {
			fmt.Fprintln(out, "Safety Meeting Job Plan Info:")
			fmt.Fprintln(out, string(payload))
		}
	}

	return nil
}
