package cli

import "strings"

func normalizeRelatedContentTypeForFilter(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	key := strings.ToLower(trimmed)
	switch key {
	case "newsletter", "newsletters":
		return "Newsletter"
	case "glossary-term", "glossary-terms", "glossaryterm", "glossaryterms":
		return "GlossaryTerm"
	case "release-note", "release-notes", "releasenote", "releasenotes":
		return "ReleaseNote"
	case "press-release", "press-releases", "pressrelease", "pressreleases":
		return "PressRelease"
	case "objective", "objectives":
		return "Objective"
	case "feature", "features":
		return "Feature"
	case "question", "questions":
		return "Question"
	default:
		return trimmed
	}
}

func normalizeRelatedContentTypeForRelationship(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	key := strings.ToLower(trimmed)
	switch key {
	case "newsletter", "newsletters":
		return "newsletters"
	case "glossary-term", "glossary-terms", "glossaryterm", "glossaryterms":
		return "glossary-terms"
	case "release-note", "release-notes", "releasenote", "releasenotes":
		return "release-notes"
	case "press-release", "press-releases", "pressrelease", "pressreleases":
		return "press-releases"
	case "objective", "objectives":
		return "objectives"
	case "feature", "features":
		return "features"
	case "question", "questions":
		return "questions"
	default:
		return trimmed
	}
}
