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

type incidentTagIncidentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentTagIncidentDetails struct {
	ID                     string   `json:"id"`
	IncidentID             string   `json:"incident_id,omitempty"`
	IncidentHeadline       string   `json:"incident_headline,omitempty"`
	IncidentStatus         string   `json:"incident_status,omitempty"`
	IncidentKind           string   `json:"incident_kind,omitempty"`
	IncidentSeverity       string   `json:"incident_severity,omitempty"`
	IncidentStartAt        string   `json:"incident_start_at,omitempty"`
	IncidentEndAt          string   `json:"incident_end_at,omitempty"`
	IncidentDescription    string   `json:"incident_description,omitempty"`
	IncidentTagID          string   `json:"incident_tag_id,omitempty"`
	IncidentTagSlug        string   `json:"incident_tag_slug,omitempty"`
	IncidentTagName        string   `json:"incident_tag_name,omitempty"`
	IncidentTagDescription string   `json:"incident_tag_description,omitempty"`
	IncidentTagKinds       []string `json:"incident_tag_kinds,omitempty"`
}

func newIncidentTagIncidentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident tag incident details",
		Long: `Show the full details of an incident tag incident link.

Output Fields:
  ID                     Resource identifier
  Incident               Incident headline and ID
  Incident Status        Incident status
  Incident Kind          Incident kind
  Incident Severity      Incident severity
  Incident Start At      Incident start timestamp
  Incident End At        Incident end timestamp
  Incident Description   Incident description
  Incident Tag           Incident tag name/slug and ID
  Incident Tag Slug      Incident tag slug
  Incident Tag Name      Incident tag name
  Incident Tag Description  Incident tag description
  Incident Tag Kinds     Incident tag kinds

Arguments:
  <id>  The incident tag incident ID (required).`,
		Example: `  # Show an incident tag incident link
  xbe view incident-tag-incidents show 123

  # Output as JSON
  xbe view incident-tag-incidents show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentTagIncidentsShow,
	}
	initIncidentTagIncidentsShowFlags(cmd)
	return cmd
}

func init() {
	incidentTagIncidentsCmd.AddCommand(newIncidentTagIncidentsShowCmd())
}

func initIncidentTagIncidentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentTagIncidentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIncidentTagIncidentsShowOptions(cmd)
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
		return fmt.Errorf("incident tag incident id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-tag-incidents]", "incident,incident-tag")
	query.Set("include", "incident,incident-tag")
	query.Set("fields[incidents]", "headline,status,kind,severity,start-at,end-at,description")
	query.Set("fields[incident-tags]", "slug,name,description,kinds")

	body, _, err := client.Get(cmd.Context(), "/v1/incident-tag-incidents/"+id, query)
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

	details := buildIncidentTagIncidentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentTagIncidentDetails(cmd, details)
}

func parseIncidentTagIncidentsShowOptions(cmd *cobra.Command) (incidentTagIncidentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentTagIncidentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentTagIncidentDetails(resp jsonAPISingleResponse) incidentTagIncidentDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := incidentTagIncidentDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["incident"]; ok && rel.Data != nil {
		details.IncidentID = rel.Data.ID
		if incident, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := incident.Attributes
			details.IncidentHeadline = stringAttr(attrs, "headline")
			details.IncidentStatus = stringAttr(attrs, "status")
			details.IncidentKind = stringAttr(attrs, "kind")
			details.IncidentSeverity = stringAttr(attrs, "severity")
			details.IncidentStartAt = stringAttr(attrs, "start-at")
			details.IncidentEndAt = stringAttr(attrs, "end-at")
			details.IncidentDescription = stringAttr(attrs, "description")
		}
	}

	if rel, ok := resp.Data.Relationships["incident-tag"]; ok && rel.Data != nil {
		details.IncidentTagID = rel.Data.ID
		if tag, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := tag.Attributes
			details.IncidentTagSlug = stringAttr(attrs, "slug")
			details.IncidentTagName = stringAttr(attrs, "name")
			details.IncidentTagDescription = stringAttr(attrs, "description")
			details.IncidentTagKinds = stringSliceAttr(attrs, "kinds")
		}
	}

	return details
}

func renderIncidentTagIncidentDetails(cmd *cobra.Command, details incidentTagIncidentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.IncidentID != "" || details.IncidentHeadline != "" {
		label := ""
		if details.IncidentHeadline != "" {
			label = details.IncidentHeadline
		}
		fmt.Fprintf(out, "Incident: %s\n", formatRelated(label, details.IncidentID))
	}
	if details.IncidentStatus != "" {
		fmt.Fprintf(out, "Incident Status: %s\n", details.IncidentStatus)
	}
	if details.IncidentKind != "" {
		fmt.Fprintf(out, "Incident Kind: %s\n", details.IncidentKind)
	}
	if details.IncidentSeverity != "" {
		fmt.Fprintf(out, "Incident Severity: %s\n", details.IncidentSeverity)
	}
	if details.IncidentStartAt != "" {
		fmt.Fprintf(out, "Incident Start At: %s\n", details.IncidentStartAt)
	}
	if details.IncidentEndAt != "" {
		fmt.Fprintf(out, "Incident End At: %s\n", details.IncidentEndAt)
	}
	if details.IncidentDescription != "" {
		fmt.Fprintf(out, "Incident Description: %s\n", details.IncidentDescription)
	}

	if details.IncidentTagID != "" || details.IncidentTagName != "" || details.IncidentTagSlug != "" {
		label := firstNonEmpty(details.IncidentTagName, details.IncidentTagSlug)
		fmt.Fprintf(out, "Incident Tag: %s\n", formatRelated(label, details.IncidentTagID))
	}
	if details.IncidentTagName != "" {
		fmt.Fprintf(out, "Incident Tag Name: %s\n", details.IncidentTagName)
	}
	if details.IncidentTagSlug != "" {
		fmt.Fprintf(out, "Incident Tag Slug: %s\n", details.IncidentTagSlug)
	}
	if details.IncidentTagDescription != "" {
		fmt.Fprintf(out, "Incident Tag Description: %s\n", details.IncidentTagDescription)
	}
	if len(details.IncidentTagKinds) > 0 {
		fmt.Fprintf(out, "Incident Tag Kinds: %s\n", strings.Join(details.IncidentTagKinds, ", "))
	}

	return nil
}
