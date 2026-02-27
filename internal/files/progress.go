package files

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

// ProgressReader wraps an io.Reader to track progress
type ProgressReader struct {
	reader   io.Reader
	progress *progressbar.ProgressBar
	total    int64
	current  int64
}

// NewProgressReader creates a new progress tracking reader
func NewProgressReader(reader io.Reader, total int64, description string, quiet bool) io.Reader {
	if quiet {
		return reader
	}

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionThrottle(100),
		progressbar.OptionFullWidth(),
		progressbar.OptionClearOnFinish(),
	)

	return &ProgressReader{
		reader:   reader,
		progress: bar,
		total:    total,
	}
}

// Read implements io.Reader
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.current += int64(n)
		_ = pr.progress.Add(n)
	}
	return n, err
}

// ProgressWriter wraps an io.Writer to track progress
type ProgressWriter struct {
	writer   io.Writer
	progress *progressbar.ProgressBar
	total    int64
	current  int64
}

// NewProgressWriter creates a new progress tracking writer
func NewProgressWriter(writer io.Writer, total int64, description string, quiet bool) io.Writer {
	if quiet {
		return writer
	}

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionThrottle(100),
		progressbar.OptionFullWidth(),
		progressbar.OptionClearOnFinish(),
	)

	return &ProgressWriter{
		writer:   writer,
		progress: bar,
		total:    total,
	}
}

// Write implements io.Writer
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.current += int64(n)
		_ = pw.progress.Add(n)
	}
	return n, err
}

// Close completes the progress bar
func (pw *ProgressWriter) Close() error {
	if pw.progress != nil {
		_ = pw.progress.Finish()
	}
	return nil
}
