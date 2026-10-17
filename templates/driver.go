package templates

import (
	"strings"

	"muttley/sqlite/entities"
)

func DriverName(user entities.User) string {
	if name := strings.TrimSpace(user.DisplayName); name != "" {
		return name
	}
	if name := joinedName(user.FirstName, user.LastName); name != "" {
		return name
	}
	return emailLocalPart(user.Email)
}

func cardTitle(user entities.GetUsersBySearchTermRow) string {
	if name := strings.TrimSpace(user.DisplayName); name != "" {
		return name
	}
	return joinedName(user.FirstName, user.LastName)
}

func joinedName(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	switch {
	case first == "" && last == "":
		return ""
	case first == "":
		return last
	case last == "":
		return first
	default:
		return first + " " + last
	}
}

func emailLocalPart(email string) string {
	local, _, ok := strings.Cut(email, "@")
	if !ok {
		return ""
	}
	return local
}
