package fileupload

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"muttley/eventresult"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func TestProcessFileRejectsBadUpload(t *testing.T) {
	db := testdb.Open(t)
	current := createUploadUser(t, db)

	noFile, noFileType := uploadMultipart(t, nil)
	empty, emptyType := uploadMultipart(t, [][]byte{{}})
	two, twoType := uploadMultipart(t, [][]byte{[]byte("one"), []byte("two")})
	tooBig, tooBigType := uploadMultipart(t, [][]byte{bytes.Repeat([]byte("a"), (2<<20)+1)})
	notEmail, notEmailType := uploadMultipart(t, [][]byte{[]byte("this is not an email")})

	tests := []struct {
		name        string
		body        *bytes.Buffer
		contentType string
		want        string
	}{
		{name: "no file", body: noFile, contentType: noFileType, want: "Choose one results file."},
		{name: "empty file", body: empty, contentType: emptyType, want: "Choose one results file."},
		{name: "two files", body: two, contentType: twoType, want: "Choose one results file."},
		{name: "too large", body: tooBig, contentType: tooBigType, want: "That file is too large."},
		{name: "not email", body: notEmail, contentType: notEmailType, want: "This file does not look like a results email."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := postUpload(t, db, current, tc.body, tc.contentType)
			body := rec.Body.String()
			if !strings.Contains(body, tc.want) {
				t.Fatalf("body = %s, missing %q", body, tc.want)
			}
			if !strings.Contains(body, "upload-file-notice") || !strings.Contains(body, `name="file-location"`) {
				t.Fatalf("body = %s, want notice and track select", body)
			}
			if strings.Contains(body, "read message:") || strings.Contains(body, "driver position:") {
				t.Fatalf("body leaked parser detail: %s", body)
			}
		})
	}
}

func TestUploadFileRendersWithoutNotice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdb.Open(t)
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))
	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	router.GET("/upload", handler.UploadFile)

	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, "upload-file-notice") {
		t.Fatalf("clean upload page included a notice: %s", body)
	}
	if !strings.Contains(body, `name="file-location"`) {
		t.Fatalf("upload form missing track select: %s", body)
	}
}

func uploadMultipart(t *testing.T, files [][]byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for i, content := range files {
		part, err := writer.CreateFormFile("file", fmt.Sprintf("result-%d.eml", i))
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("write file part: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	return &buf, writer.FormDataContentType()
}

func postUpload(t *testing.T, db *sql.DB, current entities.User, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))
	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	router.POST("/upload/process", func(c *gin.Context) {
		c.Set("currentUser", current)
		c.Next()
	}, handler.ProcessFile)

	req := httptest.NewRequest(http.MethodPost, "/upload/process", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

type testTemplRender struct {
	Code int
	Data templ.Component
}

func (t testTemplRender) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	w.WriteHeader(t.Code)
	if t.Data != nil {
		return t.Data.Render(context.Background(), w)
	}
	return nil
}

func (t testTemplRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (t *testTemplRender) Instance(name string, data interface{}) render.Render {
	if component, ok := data.(templ.Component); ok {
		return &testTemplRender{Code: http.StatusOK, Data: component}
	}
	return nil
}
