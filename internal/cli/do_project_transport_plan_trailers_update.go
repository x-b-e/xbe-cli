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

type doProjectTransportPlanTrailersUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Status       string
	SegmentStart string
	SegmentEnd   string
	Trailer      string
}

func newDoProjectTransportPlanTrailersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan trailer assignment",
		Long: `Update a project transport plan trailer assignment.

Optional:
  --status         Assignment status (editing, active)
  --segment-start  Segment start ID
  --segment-end    Segment end ID
  --trailer        Trailer ID`,
		Example: `  # Update status
  xbe do project-transport-plan-trailers update 123 --status editing

  # Update trailer assignment
  xbe do project-transport-plan-trailers update 123 --trailer 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanTrailersUpdate,
	}
	initDoProjectTransportPlanTrailersUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanTrailersCmd.AddCommand(newDoProjectTransportPlanTrailersUpdateCmd())
}

func initDoProjectTransportPlanTrailersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Assignment status (editing, active)")
	cmd.Flags().String("segment-start", "", "Segment start ID")
	cmd.Flags().String("segment-end", "", "Segment end ID")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanTrailersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanTrailersUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{}
	if opts.SegmentStart != "" {
		relationships["segment-start"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentStart,
			},
		}
	}
	if opts.SegmentEnd != "" {
		relationships["segment-end"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentEnd,
			},
		}
	}
	if opts.Trailer != "" {
		relationships["trailer"] = map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.Trailer,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-transport-plan-trailers",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-trailers/"+opts.ID, jsonBody)
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

	details := buildProjectTransportPlanTrailerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan trailer %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanTrailersUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanTrailersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	segmentStart, _ := cmd.Flags().GetString("segment-start")
	segmentEnd, _ := cmd.Flags().GetString("segment-end")
	trailer, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanTrailersUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Status:       status,
		SegmentStart: segmentStart,
		SegmentEnd:   segmentEnd,
		Trailer:      trailer,
	}, nil
}
