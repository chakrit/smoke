package checks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

type httpRequest struct {
	method string
	url    string
	// TODO: Support headers and body. Might need to extend the specs format to allow more
	//   complex check specs.
}

func (c httpRequest) Spec() string {
	return strings.ToUpper(c.method) + " " + c.url
}

func (c httpRequest) Prepare(cmd *exec.Cmd) error {
	return nil
}

func (c httpRequest) Collect(cmd *exec.Cmd) ([]byte, error) {
	req, err := http.NewRequest(c.method, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// buf := strings.Builder{}
	buf := &bytes.Buffer{}
	buf.WriteString(resp.Status)
	buf.WriteString("\n")
	// TODO: Allow selecting headers to record, as recording all headers is not practical
	//   because some headers like ETag or Date will always change.

	buf.WriteString("\n")
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	} else {
		return buf.Bytes(), nil
	}
}

func (c httpRequest) Format(buf []byte) ([]string, error) {
	return strings.Split(string(buf), "\n"), nil
}
