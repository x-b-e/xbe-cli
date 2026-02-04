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

type commitmentItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type commitmentItemDetails struct {
	ID                         string   `json:"id"`
	Label                      string   `json:"label,omitempty"`
	Status                     string   `json:"status,omitempty"`
	StartOn                    string   `json:"start_on,omitempty"`
	EndOn                      string   `json:"end_on,omitempty"`
	Years                      []string `json:"years,omitempty"`
	Months                     []string `json:"months,omitempty"`
	Weeks                      []string `json:"weeks,omitempty"`
	DaysOfWeek                 []string `json:"days_of_week,omitempty"`
	TimesOfDay                 []string `json:"times_of_day,omitempty"`
	AdjustmentSequence         string   `json:"adjustment_sequence,omitempty"`
	AdjustmentSequencePosition string   `json:"adjustment_sequence_position,omitempty"`
	AdjustmentCoefficient      float64  `json:"adjustment_coefficient,omitempty"`
	AdjustmentConstant         any      `json:"adjustment_constant,omitempty"`
	AdjustmentInput            any      `json:"adjustment_input,omitempty"`
	CommitmentType             string   `json:"commitment_type,omitempty"`
	CommitmentID               string   `json:"commitment_id,omitempty"`
}

func newCommitmentItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show commitment item details",
		Long: `Show the full details of a commitment item.

Output Fields:
  ID
  Commitment (type/id)
  Label
  Status
  Start On
  End On
  Years
  Months
  Weeks
  Days Of Week
  Times Of Day
  Adjustment Sequence Position
  Adjustment Sequence
  Adjustment Coefficient
  Adjustment Constant
  Adjustment Input

Arguments:
  <id>    The commitment item ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a commitment item
  xbe view commitment-items show 123

  # Output as JSON
  xbe view commitment-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommitmentItemsShow,
	}
	initCommitmentItemsShowFlags(cmd)
	return cmd
}

func init() {
	commitmentItemsCmd.AddCommand(newCommitmentItemsShowCmd())
}

func initCommitmentItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentItemsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseCommitmentItemsShowOptions(cmd)
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
		return fmt.Errorf("commitment item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[commitment-items]", "start-on,end-on,years,months,weeks,days-of-week,times-of-day,adjustment-sequence,adjustment-sequence-position,label,status,adjustment-constant,adjustment-coefficient,adjustment-input,commitment")

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-items/"+id, query)
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

	details := buildCommitmentItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommitmentItemDetails(cmd, details)
}

func parseCommitmentItemsShowOptions(cmd *cobra.Command) (commitmentItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommitmentItemDetails(resp jsonAPISingleResponse) commitmentItemDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := commitmentItemDetails{
		ID:                         resource.ID,
		Label:                      stringAttr(attrs, "label"),
		Status:                     stringAttr(attrs, "status"),
		StartOn:                    formatDate(stringAttr(attrs, "start-on")),
		EndOn:                      formatDate(stringAttr(attrs, "end-on")),
		Years:                      stringSliceAttr(attrs, "years"),
		Months:                     stringSliceAttr(attrs, "months"),
		Weeks:                      stringSliceAttr(attrs, "weeks"),
		DaysOfWeek:                 stringSliceAttr(attrs, "days-of-week"),
		TimesOfDay:                 stringSliceAttr(attrs, "times-of-day"),
		AdjustmentSequence:         stringAttr(attrs, "adjustment-sequence"),
		AdjustmentSequencePosition: stringAttr(attrs, "adjustment-sequence-position"),
		AdjustmentCoefficient:      floatAttr(attrs, "adjustment-coefficient"),
	}

	if value, ok := attrs["adjustment-constant"]; ok {
		details.AdjustmentConstant = value
	}
	if value, ok := attrs["adjustment-input"]; ok {
		details.AdjustmentInput = value
	}
	if rel, ok := resource.Relationships["commitment"]; ok && rel.Data != nil {
		details.CommitmentType = rel.Data.Type
		details.CommitmentID = rel.Data.ID
	}

	return details
}

func renderCommitmentItemDetails(cmd *cobra.Command, details commitmentItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CommitmentType != "" || details.CommitmentID != "" {
		commitment := details.CommitmentType
		if details.CommitmentID != "" {
			if commitment != "" {
				commitment += "/"
			}
			commitment += details.CommitmentID
		}
		fmt.Fprintf(out, "Commitment: %s\n", commitment)
	}
	if details.Label != "" {
		fmt.Fprintf(out, "Label: %s\n", details.Label)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", details.EndOn)
	}
	if len(details.Years) > 0 {
		fmt.Fprintf(out, "Years: %s\n", strings.Join(details.Years, ", "))
	}
	if len(details.Months) > 0 {
		fmt.Fprintf(out, "Months: %s\n", strings.Join(details.Months, ", "))
	}
	if len(details.Weeks) > 0 {
		fmt.Fprintf(out, "Weeks: %s\n", strings.Join(details.Weeks, ", "))
	}
	if len(details.DaysOfWeek) > 0 {
		fmt.Fprintf(out, "Days Of Week: %s\n", strings.Join(details.DaysOfWeek, ", "))
	}
	if len(details.TimesOfDay) > 0 {
		fmt.Fprintf(out, "Times Of Day: %s\n", strings.Join(details.TimesOfDay, ", "))
	}
	if details.AdjustmentSequencePosition != "" {
		fmt.Fprintf(out, "Adjustment Sequence Position: %s\n", details.AdjustmentSequencePosition)
	}
	if details.AdjustmentSequence != "" {
		fmt.Fprintf(out, "Adjustment Sequence: %s\n", details.AdjustmentSequence)
	}
	if details.AdjustmentCoefficient != 0 {
		fmt.Fprintf(out, "Adjustment Coefficient: %s\n", formatOptionalFloat(details.AdjustmentCoefficient))
	}
	if details.AdjustmentConstant != nil {
		pretty := formatJSONValue(details.AdjustmentConstant)
		if pretty != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Adjustment Constant:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, pretty)
		}
	}
	if details.AdjustmentInput != nil {
		pretty := formatJSONValue(details.AdjustmentInput)
		if pretty != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Adjustment Input:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, pretty)
		}
	}

	return nil
}
