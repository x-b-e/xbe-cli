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

type doDevicesUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	Nickname       string
	PushableStatus string
	IsPreferred    bool
}

func newDoDevicesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a device",
		Long: `Update a device.

Note: Devices cannot be created or deleted via the API. They are created automatically
when users log in to the mobile app.

Optional flags:
  --nickname          Device nickname
  --pushable-status   Pushable status
  --is-preferred      Whether this is the user's preferred device`,
		Example: `  # Update device nickname
  xbe do devices update 123 --nickname "Work Phone"

  # Update preferred status
  xbe do devices update 123 --is-preferred true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDevicesUpdate,
	}
	initDoDevicesUpdateFlags(cmd)
	return cmd
}

func init() {
	doDevicesCmd.AddCommand(newDoDevicesUpdateCmd())
}

func initDoDevicesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("nickname", "", "Device nickname")
	cmd.Flags().String("pushable-status", "", "Pushable status")
	cmd.Flags().Bool("is-preferred", false, "Whether this is the user's preferred device")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDevicesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDevicesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("nickname") {
		attributes["nickname"] = opts.Nickname
	}
	if cmd.Flags().Changed("pushable-status") {
		attributes["pushable-status"] = opts.PushableStatus
	}
	if cmd.Flags().Changed("is-preferred") {
		attributes["is-preferred"] = opts.IsPreferred
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "devices",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/devices/"+opts.ID, jsonBody)
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

	row := buildDeviceRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated device %s\n", row.ID)
	return nil
}

func parseDoDevicesUpdateOptions(cmd *cobra.Command, args []string) (doDevicesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	nickname, _ := cmd.Flags().GetString("nickname")
	pushableStatus, _ := cmd.Flags().GetString("pushable-status")
	isPreferred, _ := cmd.Flags().GetBool("is-preferred")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDevicesUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		Nickname:       nickname,
		PushableStatus: pushableStatus,
		IsPreferred:    isPreferred,
	}, nil
}

func buildDeviceRowFromSingle(resp jsonAPISingleResponse) deviceRow {
	attrs := resp.Data.Attributes

	row := deviceRow{
		ID:                         resp.Data.ID,
		Identifier:                 stringAttr(attrs, "identifier"),
		Nickname:                   stringAttr(attrs, "nickname"),
		IsPushable:                 boolAttr(attrs, "is-pushable"),
		IsPreferred:                boolAttr(attrs, "is-preferred"),
		HasPushToken:               boolAttr(attrs, "has-push-token"),
		PushableStatus:             stringAttr(attrs, "pushable-status"),
		LastNativeAppVersion:       stringAttr(attrs, "last-native-app-version"),
		FirstDeviceLocationEventAt: stringAttr(attrs, "first-device-location-event-at"),
		LastDeviceLocationEventAt:  stringAttr(attrs, "last-device-location-event-at"),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}
