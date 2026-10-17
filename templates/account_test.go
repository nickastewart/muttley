package templates_test

import (
	"bytes"
	"context"
	"testing"

	"muttley/sqlite/entities"
	"muttley/templates"
)

func TestAccountShowsMandatoryFieldsBanner(t *testing.T) {
	var buf bytes.Buffer
	component := templates.Account(entities.User{}, "", "")
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("All fields are mandatory.")) {
		t.Fatalf("account page = %s", buf.String())
	}
}
