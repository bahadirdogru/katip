package kitap

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Metadata struct {
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PlotSummary  string    `json:"plot_summary"`
	BaseFontSize int       `json:"base_font_size,omitempty"`
	FontFamily   string    `json:"font_family,omitempty"`
}

type Revision struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	ChangeLog string    `json:"change_log"`
}

type CommentReply struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type CommentThread struct {
	ID       string         `json:"id"`
	Quote    string         `json:"quote"`
	Resolved bool           `json:"resolved"`
	Replies  []CommentReply `json:"replies"`
}

type Document struct {
	Metadata Metadata        `json:"metadata"`
	History  []Revision      `json:"history"`
	Comments []CommentThread `json:"comments"`
	Content  string          `json:"content"`
	TempDir  string          `json:"-"`
}

func initDocument() *Document {
	return &Document{
		Metadata: Metadata{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		History:  []Revision{},
		Comments: []CommentThread{},
	}
}

// NewKitap creates an empty Document structure
func NewKitap() *Document {
	return initDocument()
}

// Write saves the document object to a .kitap file (ZIP archive)
func Write(doc *Document, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	if err := addFileToZip(w, "content.md", []byte(doc.Content)); err != nil {
		return err
	}

	metaBytes, _ := json.MarshalIndent(doc.Metadata, "", "  ")
	if err := addFileToZip(w, "metadata.json", metaBytes); err != nil {
		return err
	}

	historyBytes, _ := json.MarshalIndent(doc.History, "", "  ")
	if err := addFileToZip(w, "history.json", historyBytes); err != nil {
		return err
	}

	commentsBytes, _ := json.MarshalIndent(doc.Comments, "", "  ")
	if err := addFileToZip(w, "comments.json", commentsBytes); err != nil {
		return err
	}

	return nil
}

// Read loads a Document from a .kitap file
func Read(filePath string) (*Document, error) {
	doc := initDocument()

	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open .kitap: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(rc)
		rc.Close()
		
		if err != nil {
			return nil, err
		}

		switch f.Name {
		case "content.md":
			doc.Content = string(b)
		case "metadata.json":
			json.Unmarshal(b, &doc.Metadata)
		case "history.json":
			json.Unmarshal(b, &doc.History)
		case "comments.json":
			json.Unmarshal(b, &doc.Comments)
		}
	}

	return doc, nil
}

func addFileToZip(w *zip.Writer, filename string, content []byte) error {
	f, err := w.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}
