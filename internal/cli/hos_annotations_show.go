package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type hosAnnotationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosAnnotationDetails struct {
	ID           string `json:"id"`
	AnnotationAt string `json:"annotation_at,omitempty"`
	Comment      string `json:"comment,omitempty"`
	Metadata     any    `json:"metadata,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
	HosDayID     string `json:"hos_day_id,omitempty"`
	HosEventID   string `json:"hos_event_id,omitempty"`
}

func newHosAnnotationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS annotation details",
		Long: `Show the full details of a HOS annotation.

Output Fields:
  ID
  Annotation At
  Comment
  Broker ID
  HOS Day ID
  HOS Event ID
  Metadata

Arguments:
  <id>  The annotation ID (required). You can find IDs using the list command.`,
		Example: `  # Show an annotation
  xbe view hos-annotations show 123

  # Get JSON output
  xbe view hos-annotations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosAnnotationsShow,
	}
	initHosAnnotationsShowFlags(cmd)
	return cmd
}

func init() {
	hosAnnotationsCmd.AddCommand(newHosAnnotationsShowCmd())
}

func initHosAnnotationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosAnnotationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseHosAnnotationsShowOptions(cmd)
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
		return fmt.Errorf("hos annotation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-annotations/"+id, nil)
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

	details := buildHosAnnotationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosAnnotationDetails(cmd, details)
}

func parseHosAnnotationsShowOptions(cmd *cobra.Command) (hosAnnotationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosAnnotationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosAnnotationDetails(resp jsonAPISingleResponse) hosAnnotationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := hosAnnotationDetails{
		ID:           resource.ID,
		AnnotationAt: formatDateTime(stringAttr(attrs, "annotation-at")),
		Comment:      strings.TrimSpace(stringAttr(attrs, "comment")),
		Metadata:     attrs["metadata"],
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		details.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-event"]; ok && rel.Data != nil {
		details.HosEventID = rel.Data.ID
	}

	return details
}

func renderHosAnnotationDetails(cmd *cobra.Command, details hosAnnotationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.AnnotationAt != "" {
		fmt.Fprintf(out, "Annotation At: %s\n", details.AnnotationAt)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.HosDayID != "" {
		fmt.Fprintf(out, "HOS Day ID: %s\n", details.HosDayID)
	}
	if details.HosEventID != "" {
		fmt.Fprintf(out, "HOS Event ID: %s\n", details.HosEventID)
	}

	metadata := formatHosAnnotationMetadata(details.Metadata)
	if metadata != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Metadata:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, metadata)
	}

	return nil
}

func formatHosAnnotationMetadata(value any) string {
	if value == nil {
		return ""
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(payload)
}
