// Command ocrtest ingests image files through the real pipeline and writes the
// results as CSV: one summary row per page plus one row per OCR text line with
// best-effort cell splitting.
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/ingest"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
)

var cellSep = regexp.MustCompile(`[\[\](){}|]+`)

func main() {
	outDir := os.Getenv("OUT")
	if outDir == "" {
		outDir = "out"
	}
	paths := os.Args[1:]
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ocrtest <image>... (env: OUT)")
		os.Exit(2)
	}

	svc := ingest.New(ocr.Configured(ocr.Options{}), ingest.Limits{})
	fmt.Printf("engine=%s available=%v files=%d\n", svc.Engine().Name(), svc.Engine().Available(), len(paths))

	must(os.MkdirAll(outDir, 0o755))
	pagesF, err := os.Create(filepath.Join(outDir, "pages.csv"))
	must(err)
	defer pagesF.Close()
	rowsF, err := os.Create(filepath.Join(outDir, "rows.csv"))
	must(err)
	defer rowsF.Close()

	pagesW := csv.NewWriter(pagesF)
	rowsW := csv.NewWriter(rowsF)
	defer func() { pagesW.Flush(); rowsW.Flush() }()
	must(pagesW.Write([]string{"file", "format", "width", "height", "confidence_pct", "text_chars", "error"}))
	must(rowsW.Write([]string{"file", "line", "col1", "col2", "col3", "col4", "col5", "col6"}))

	ok := 0
	var confSum float64
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: read error: %v\n", path, err)
			continue
		}
		name := filepath.Base(path)
		doc, err := svc.Ingest(context.Background(), ingest.Request{
			Name:    name,
			Content: content,
			Options: format.Options{OCR: true},
		})
		if err != nil {
			must(pagesW.Write([]string{name, "", "", "", "", "0", err.Error()}))
			continue
		}
		for _, p := range doc.Pages {
			confStr := ""
			if p.Confidence != document.ConfidenceUnknown {
				confSum += p.Confidence * 100
				confStr = strconv.FormatFloat(p.Confidence*100, 'f', 1, 64)
			}
			errStr := ""
			if p.Err != nil {
				errStr = p.Err.Code + ": " + p.Err.Message
			}
			w, h := "", ""
			if len(p.Images) > 0 {
				w, h = strconv.Itoa(p.Images[0].Width), strconv.Itoa(p.Images[0].Height)
			}
			must(pagesW.Write([]string{name, doc.Format.String(), w, h, confStr, strconv.Itoa(len(strings.TrimSpace(p.Text))), errStr}))
			for i, line := range strings.Split(p.Text, "\n") {
				cells := splitCells(line)
				if len(cells) == 0 {
					continue
				}
				rec := []string{name, strconv.Itoa(i + 1)}
				rec = append(rec, cells...)
				must(rowsW.Write(rec))
			}
		}
		ok++
	}
	fmt.Printf("%d ok, %d failed; mean confidence %.1f%%\n", ok, len(paths)-ok, confSum/float64(ok))
	fmt.Printf("wrote %s/pages.csv and %s/rows.csv\n", outDir, outDir)
}

func splitCells(line string) []string {
	var cells []string
	for _, c := range cellSep.Split(line, -1) {
		if c = strings.TrimSpace(c); c != "" {
			cells = append(cells, c)
		}
	}
	if len(cells) > 6 {
		cells = append(cells[:5], strings.Join(cells[5:], " "))
	}
	return cells
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
