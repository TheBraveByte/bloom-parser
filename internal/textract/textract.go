// Package textract runs the Python tableextract pipeline (OpenCV grid
// detection + Tesseract OCR, optional vision-LLM refinement, normalization)
// as a subprocess and returns the normalized CSV it produces.
package textract

import (
	"bytes"
	"context"
	gocsv "encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Runner executes the tableextract CLI for uploaded images.
type Runner struct {
	// Python is the interpreter used to run the module
	// (env TABLEEXTRACT_PYTHON, default "python3").
	Python string
	// Dir is the working directory that contains the tableextract package
	// (env TABLEEXTRACT_DIR, default "tableextract"). Ignored if the path
	// does not exist, e.g. when the package is pip-installed into Python.
	Dir string
	// Timeout bounds the pipeline per file. Refinement can be slow, so when
	// Options.Refine is set the effective deadline grows to ~6x Timeout and
	// refine itself is bounded to ~4x Timeout via --refine-budget.
	Timeout time.Duration
	// Workers is the max number of files processed concurrently.
	Workers int
}

// NewRunner builds a Runner from the environment.
func NewRunner() *Runner {
	r := &Runner{
		Python:  envOr("TABLEEXTRACT_PYTHON", "python3"),
		Dir:     envOr("TABLEEXTRACT_DIR", "tableextract"),
		Timeout: 3 * time.Minute,
		Workers: 4,
	}
	if d, err := time.ParseDuration(os.Getenv("TABLEEXTRACT_TIMEOUT")); err == nil {
		r.Timeout = d
	}
	return r
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// Options controls one pipeline run.
type Options struct {
	// Refine runs the vision-LLM pass between extract and normalize.
	// It needs NVIDIA_API_KEY and is much slower.
	Refine bool
}

// File is one uploaded image.
type File struct {
	Name string
	Data []byte
}

// FileStatus reports the per-file outcome inside a bulk Result.
type FileStatus struct {
	File    string `json:"file"`
	Rows    int    `json:"rows"`
	Flagged int    `json:"flagged"`
	Error   string `json:"error,omitempty"`
}

// Result is the merged outcome of a bulk run.
type Result struct {
	CSV     string       `json:"csv"`
	Rows    int          `json:"rows"`
	Flagged int          `json:"flagged"`
	Files   []FileStatus `json:"files"`
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func safeName(name string) string {
	name = unsafeName.ReplaceAllString(filepath.Base(name), "_")
	if name == "" || name == "." {
		return "input"
	}
	return name
}

// RunAll processes files concurrently (bounded by Workers) and merges their
// normalized CSVs. A failed file is reported in Result.Files and does not
// fail the whole run.
func (r *Runner) RunAll(ctx context.Context, files []File, opts Options) *Result {
	type outcome struct {
		name       string
		rows       []string
		facts, flg int
		err        error
	}
	jobs := make(chan File)
	outs := make(chan outcome, len(files))
	var wg sync.WaitGroup

	workers := r.Workers
	if n := len(files); workers < 1 || workers > n {
		workers = max(n, 1)
	}
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				rows, facts, flg, err := r.runOne(ctx, f, opts)
				outs <- outcome{f.Name, rows, facts, flg, err}
			}
		}()
	}
	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	go func() { wg.Wait(); close(outs) }()

	var buf bytes.Buffer
	buf.WriteString("file,row,section,line_item,period,value,raw,flag\r\n")
	res := &Result{}
	for o := range outs {
		st := FileStatus{File: o.name}
		if o.err != nil {
			st.Error = o.err.Error()
		} else {
			st.Rows, st.Flagged = o.facts, o.flg
			res.Rows += o.facts
			res.Flagged += o.flg
			for _, row := range o.rows {
				buf.WriteString(row)
				buf.WriteString("\r\n")
			}
		}
		res.Files = append(res.Files, st)
	}
	res.CSV = buf.String()
	return res
}

func (r *Runner) runOne(ctx context.Context, f File, opts Options) (rows []string, facts, flagged int, err error) {
	tmp, err := os.MkdirTemp("", "tableextract-")
	if err != nil {
		return nil, 0, 0, fmt.Errorf("tmpdir: %w", err)
	}
	defer os.RemoveAll(tmp)

	img := filepath.Join(tmp, safeName(f.Name))
	if err := os.WriteFile(img, f.Data, 0o600); err != nil {
		return nil, 0, 0, fmt.Errorf("write upload: %w", err)
	}
	out := filepath.Join(tmp, "out")

	// One subprocess runs extract -> (refine) -> normalize. The old
	// three-stage version paid the interpreter + cv2 startup per stage.
	// Refine stays best-effort inside the pipeline: --refine-budget keeps it
	// within ~4x Timeout and a refine failure still produces normalized CSV.
	args := []string{"pipeline", out, img}
	deadline := r.Timeout
	if opts.Refine {
		args = append(args, "--refine",
			fmt.Sprintf("--refine-budget=%d", int((r.Timeout*4).Seconds())))
		deadline = r.Timeout * 6
	}
	if err := r.stage(ctx, deadline, args...); err != nil {
		return nil, 0, 0, err
	}

	norm, err := os.ReadFile(filepath.Join(out, "normalized.csv"))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read normalized.csv: %w", err)
	}
	recs, err := gocsv.NewReader(bytes.NewReader(norm)).ReadAll()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("parse normalized.csv: %w", err)
	}
	// Re-encode with a single shared writer (one allocation run) rather than
	// a fresh csv.Writer per row.
	var buf bytes.Buffer
	w := gocsv.NewWriter(&buf)
	for _, rec := range recs[1:] {
		facts++
		if len(rec) > 7 && rec[7] != "" {
			flagged++
		}
		_ = w.Write(rec)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, 0, 0, fmt.Errorf("encode csv: %w", err)
	}
	if body := strings.TrimRight(buf.String(), "\n"); body != "" {
		rows = strings.Split(body, "\n")
	}
	return rows, facts, flagged, nil
}

// stage runs the pipeline subprocess under a timeout derived from ctx.
func (r *Runner) stage(ctx context.Context, d time.Duration, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return r.exec(ctx, args...)
}

func (r *Runner) exec(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, r.Python, append([]string{"-m", "tableextract"}, args...)...)
	if st, err := os.Stat(r.Dir); err == nil && st.IsDir() {
		cmd.Dir = r.Dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tableextract %s: %w: %s", args[0], err, stderr.String())
	}
	return nil
}
