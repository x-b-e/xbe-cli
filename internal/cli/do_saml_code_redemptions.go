package cli

import "github.com/spf13/cobra"

var doSamlCodeRedemptionsCmd = &cobra.Command{
	Use:   "saml-code-redemptions",
	Short: "Redeem SAML login codes",
	Long: `Redeem SAML login codes.

SAML code redemptions exchange a SAML login code for an XBE auth token.

Commands:
  create    Redeem a SAML login code`,
	Example: `  # Redeem a SAML login code
  xbe do saml-code-redemptions create --code "saml_code_value"

  # JSON output
  xbe do saml-code-redemptions create --code "saml_code_value" --json`,
}

func init() {
	doCmd.AddCommand(doSamlCodeRedemptionsCmd)
}
