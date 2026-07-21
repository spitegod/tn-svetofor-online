package service

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestValidateSystemDocumentAttachmentRejectsSpoofedPDF(t *testing.T) {
	if _, err := validateSystemDocumentAttachment(".pdf", []byte("<svg></svg>")); err == nil {
		t.Fatal("expected spoofed PDF to be rejected")
	}
}

func TestValidateSystemDocumentAttachmentAcceptsPDF(t *testing.T) {
	contentType, err := validateSystemDocumentAttachment(".pdf", []byte("%PDF-1.7\n"))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "application/pdf" {
		t.Fatalf("content type = %q", contentType)
	}
}

func TestValidateSystemDocumentAttachmentAcceptsDOCX(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for _, name := range []string{"[Content_Types].xml", "word/document.xml"} {
		file, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte("<xml/>")); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := validateSystemDocumentAttachment(".docx", buffer.Bytes()); err != nil {
		t.Fatal(err)
	}
}
