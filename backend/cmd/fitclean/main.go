// fitclean scans a directory of FIT files, moves all non-activity files
// (monitoring, weight, segments, unknown types, etc.) into a sub-directory,
// and prints a summary report.
//
// Usage:
//
//	go run ./cmd/fitclean [dir]
//
// dir defaults to ../.fit
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tormoder/fit"
)

func main() {
	dir := "../.fit"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read directory %q: %v\n", dir, err)
		os.Exit(1)
	}

	archiveDir := filepath.Join(dir, "non-activities")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create archive dir: %v\n", err)
		os.Exit(1)
	}

	type counts struct{ kept, moved, errored int }
	var c counts
	movedByReason := map[string]int{}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".fit") {
			continue
		}

		src := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			fmt.Printf("  ERROR  read  %s: %v\n", e.Name(), err)
			c.errored++
			continue
		}

		reason := classifyNonActivity(data)
		if reason == "" {
			// It's an activity — leave it in place.
			c.kept++
			continue
		}

		dst := filepath.Join(archiveDir, e.Name())
		if err := os.Rename(src, dst); err != nil {
			fmt.Printf("  ERROR  move  %s: %v\n", e.Name(), err)
			c.errored++
			continue
		}
		c.moved++
		movedByReason[reason]++
	}

	fmt.Println("\n=== fitclean report ===")
	fmt.Printf("  kept (activities)  %d\n", c.kept)
	fmt.Printf("  moved              %d → %s\n", c.moved, archiveDir)
	if c.errored > 0 {
		fmt.Printf("  errors             %d\n", c.errored)
	}
	fmt.Println("\n  moved by reason:")
	for reason, n := range movedByReason {
		fmt.Printf("    %-40s %d\n", reason, n)
	}
}

// classifyNonActivity returns a human-readable reason if the file should be
// moved out, or "" if it is a genuine activity file.
func classifyNonActivity(data []byte) string {
	f, err := fit.Decode(bytes.NewReader(data))
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unknown file type"):
			// Extract the type token for reporting (e.g. "FileType(44)")
			tok := "unknown_type"
			if i := strings.Index(msg, "FileType("); i >= 0 {
				end := strings.Index(msg[i:], ")")
				if end >= 0 {
					tok = msg[i : i+end+1]
				}
			}
			return "decode_error: " + tok
		case strings.Contains(msg, "was not for file_id"):
			return "decode_error: unexpected_first_message"
		default:
			// Genuinely corrupt / truncated — move out so they don't clog imports.
			return "decode_error: corrupt"
		}
	}

	if f.Type() == fit.FileTypeActivity {
		return "" // keep
	}
	return "non_activity: " + f.Type().String()
}
