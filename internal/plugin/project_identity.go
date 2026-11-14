package plugin

import "regexp"

var validProjectID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

func isValidProjectID(projectID string) bool {
	return validProjectID.MatchString(projectID)
}
