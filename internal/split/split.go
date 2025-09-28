// Package split provides functions to split data into multiple documents.
package split

import (
	"bufio"
	"bytes"
	"strings"
)

// YAML splits a YAML stream into raw documents,
// using lines that match `---` + optional whitespace as separators.
func YAML(data []byte) [][]byte {
	var docs [][]byte

	scanner := bufio.NewScanner(bytes.NewReader(data))

	offset := 0
	prev := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			chunk := bytes.TrimSpace(data[prev:offset])
			if len(chunk) > 0 {
				docs = append(docs, chunk)
			}

			prev = offset + len(scanner.Bytes()) + 1
		}

		offset += len(scanner.Bytes()) + 1
	}

	if prev < len(data) {
		chunk := bytes.TrimSpace(data[prev:])
		if len(chunk) > 0 {
			docs = append(docs, chunk)
		}
	}

	return docs
}
