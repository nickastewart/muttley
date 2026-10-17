package templates_test

import (
	"testing"

	"muttley/sqlite/entities"
	"muttley/templates"
)

func TestDriverName(t *testing.T) {
	if got := templates.DriverName(entities.User{
		DisplayName: "Apex",
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       "ada@example.com",
	}); got != "Apex" {
		t.Fatalf("DriverName = %q, want Apex", got)
	}

	if got := templates.DriverName(entities.User{
		FirstName: " Ada ",
		LastName:  " Lovelace ",
		Email:     "ada@example.com",
	}); got != "Ada Lovelace" {
		t.Fatalf("DriverName = %q, want Ada Lovelace", got)
	}

	if got := templates.DriverName(entities.User{Email: "ada@example.com"}); got != "ada" {
		t.Fatalf("DriverName = %q, want ada", got)
	}
	if got := templates.DriverName(entities.User{Email: "ada@example.com"}); got == "ada@example.com" {
		t.Fatalf("DriverName rendered the full address")
	}
}
