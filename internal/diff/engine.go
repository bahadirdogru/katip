package diff

import (
	"strings"

	difflib "github.com/sergi/go-diff/diffmatchpatch"
)

type Engine struct {
	dmp *difflib.DiffMatchPatch
}

type DiffChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewEngine() *Engine {
	return &Engine{
		dmp: difflib.New(),
	}
}

func (e *Engine) ComputeDiff(original, improved string) []DiffChunk {
	diffs := e.dmp.DiffMain(original, improved, false)
	diffs = e.dmp.DiffCleanupSemantic(diffs)

	chunks := make([]DiffChunk, 0, len(diffs))
	for _, d := range diffs {
		if d.Text == "" {
			continue
		}
		var t string
		switch d.Type {
		case difflib.DiffEqual:
			t = "equal"
		case difflib.DiffDelete:
			t = "delete"
		case difflib.DiffInsert:
			t = "insert"
		}
		chunks = append(chunks, DiffChunk{Type: t, Text: d.Text})
	}

	return chunks
}

func (e *Engine) ComputeWordDiff(original, improved string) []DiffChunk {
	origWords := strings.Fields(original)
	newWords := strings.Fields(improved)

	origText := strings.Join(origWords, "\n")
	newText := strings.Join(newWords, "\n")

	diffs := e.dmp.DiffMain(origText, newText, false)
	diffs = e.dmp.DiffCleanupSemantic(diffs)

	chunks := make([]DiffChunk, 0, len(diffs))
	for _, d := range diffs {
		text := strings.ReplaceAll(d.Text, "\n", " ")
		if text == "" {
			continue
		}
		var t string
		switch d.Type {
		case difflib.DiffEqual:
			t = "equal"
		case difflib.DiffDelete:
			t = "delete"
		case difflib.DiffInsert:
			t = "insert"
		}
		chunks = append(chunks, DiffChunk{Type: t, Text: text})
	}

	return chunks
}
