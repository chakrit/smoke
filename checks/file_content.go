package checks

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type fileContent struct {
	glob string
}

func (c fileContent) Spec() string {
	return c.glob
}

func (c fileContent) Prepare(cmd *exec.Cmd) error {
	return nil
}

func (c fileContent) Collect(cmd *exec.Cmd) ([]byte, error) {
	matches, err := filepath.Glob(c.glob)
	if err != nil {
		return nil, fmt.Errorf("file: bad glob or i/o error `%s`", c.glob)
	}

	buf := &bytes.Buffer{}
	w := tar.NewWriter(buf)

	sort.Sort(sort.StringSlice(matches))
	for _, filename := range matches {
		if info, err := os.Stat(filename); err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		} else if hdr, err := tar.FileInfoHeader(info, ""); err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		} else if err := w.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		}

		if file, err := os.Open(filename); err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		} else if _, err = io.Copy(w, file); err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		} else if err = file.Close(); err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return nil, fmt.Errorf("file: i/o: %w", err)
	} else if err := w.Close(); err != nil {
		return nil, fmt.Errorf("file: i/o: %w", err)
	}

	return buf.Bytes(), nil
}

func (c fileContent) Format(buf []byte) ([]string, error) {
	var lines []string
	r := tar.NewReader(bytes.NewBuffer(buf))

	for hdr, err := r.Next(); err != io.EOF; hdr, err = r.Next() {
		buf, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("file: i/o: %w", err)
		}

		buf = bytes.Trim(buf, "\r\n")
		lines = append(lines, "-----BEGIN "+hdr.Name+"-----")
		lines = append(lines, strings.Split(string(buf), "\n")...)
		lines = append(lines, "-----END "+hdr.Name+"-----", "")
	}

	return lines, nil
}
