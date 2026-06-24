package forms_test

import (
	"bytes"
	"github.com/pkshahid/JanGo/forms"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"testing"
)

func createTestFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(32 << 20)
	if err != nil {
		t.Fatal(err)
	}

	return req.MultipartForm.File["file"][0]
}

func TestFileField(t *testing.T) {
	field := forms.FileField{
		BaseField:    forms.BaseField{IsRequired: true},
		MaxBytes:     100,
		AllowedTypes: []string{"text/plain", "application/octet-stream"},
	}

	_, err := field.Clean(nil)
	if err == nil {
		t.Error("Expected error for required field")
	}

	// Test size limit
	largeContent := bytes.Repeat([]byte("a"), 150)
	headerLarge := createTestFileHeader(t, "large.txt", largeContent)
	_, err = field.Clean(headerLarge)
	if err == nil || err.Error() != "file size exceeds maximum allowed" {
		t.Errorf("Expected size error, got %v", err)
	}

	// Test allowed types
	headerInvalidType := createTestFileHeader(t, "test.html", []byte("<html><body>Hello</body></html>"))
	_, err = field.Clean(headerInvalidType)
	if err == nil || err.Error() != "file type text/html; charset=utf-8 is not allowed" {
		t.Errorf("Expected type error, got %v", err)
	}

	// Test success
	headerValid := createTestFileHeader(t, "test.txt", []byte("hello world"))
	res, err := field.Clean(headerValid)
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
	if res.(*multipart.FileHeader).Filename != "test.txt" {
		t.Errorf("Expected filename test.txt")
	}
}

func TestImageField(t *testing.T) {
	field := forms.ImageField{}

	// Test invalid image
	headerInvalid := createTestFileHeader(t, "test.txt", []byte("not an image"))
	_, err := field.Clean(headerInvalid)
	if err == nil || err.Error() != "uploaded file is not a valid image" {
		t.Errorf("Expected image error, got %v", err)
	}

	// Test valid image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	png.Encode(&buf, img)

	headerValid := createTestFileHeader(t, "test.png", buf.Bytes())
	res, err := field.Clean(headerValid)
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
	if res.(*multipart.FileHeader).Filename != "test.png" {
		t.Errorf("Expected filename test.png")
	}
}
