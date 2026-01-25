package cli

type userSearchRow struct {
	ID                string `json:"id"`
	ContactMethod     string `json:"contact_method,omitempty"`
	ContactValue      string `json:"contact_value,omitempty"`
	OnlyAdminOrMember bool   `json:"only_admin_or_member,omitempty"`
	MatchingUserID    string `json:"matching_user_id,omitempty"`
	MatchingUserName  string `json:"matching_user_name,omitempty"`
	MatchingUserEmail string `json:"matching_user_email,omitempty"`
}

func buildUserSearchRows(resp jsonAPIResponse) []userSearchRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	rows := make([]userSearchRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildUserSearchRow(resource, included))
	}
	return rows
}

func buildUserSearchRow(resource jsonAPIResource, included map[string]jsonAPIResource) userSearchRow {
	attrs := resource.Attributes
	row := userSearchRow{
		ID:            resource.ID,
		ContactMethod: firstNonEmpty(stringAttr(attrs, "contact-method"), stringAttr(attrs, "contact_method")),
		ContactValue:  firstNonEmpty(stringAttr(attrs, "contact-value"), stringAttr(attrs, "contact_value")),
	}

	onlyAdminOrMember := boolAttr(attrs, "only-admin-or-member")
	if !onlyAdminOrMember {
		onlyAdminOrMember = boolAttr(attrs, "only_admin_or_member")
	}
	row.OnlyAdminOrMember = onlyAdminOrMember

	if rel, ok := resource.Relationships["matching-user"]; ok && rel.Data != nil {
		row.MatchingUserID = rel.Data.ID
		applyMatchingUserDetails(&row, rel.Data.Type, rel.Data.ID, included)
	} else if rel, ok := resource.Relationships["matching_user"]; ok && rel.Data != nil {
		row.MatchingUserID = rel.Data.ID
		applyMatchingUserDetails(&row, rel.Data.Type, rel.Data.ID, included)
	}

	return row
}

func applyMatchingUserDetails(row *userSearchRow, resourceType string, resourceID string, included map[string]jsonAPIResource) {
	user, ok := included[resourceKey(resourceType, resourceID)]
	if !ok {
		return
	}
	row.MatchingUserName = stringAttr(user.Attributes, "name")
	row.MatchingUserEmail = stringAttr(user.Attributes, "email-address")
}

func userSearchRowFromSingle(resp jsonAPISingleResponse) userSearchRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	return buildUserSearchRow(resp.Data, included)
}

func userSearchMatchingLabel(row userSearchRow) string {
	label := firstNonEmpty(row.MatchingUserName, row.MatchingUserEmail, row.MatchingUserID)
	if label == "" {
		return "-"
	}
	return label
}

func userSearchContactValueLabel(row userSearchRow) string {
	if row.ContactValue == "" {
		return "-"
	}
	return row.ContactValue
}

func userSearchOnlyAdminLabel(row userSearchRow) string {
	if row.OnlyAdminOrMember {
		return "true"
	}
	return "false"
}
