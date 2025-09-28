package variables

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v4"
)

// Header represents the header section of an alias file.
type Header struct {
	// Values contains predefined variables to use for templating.
	Values Variables `yaml:"values,omitempty"`
	// Exclude indicates whether this file should be excluded from processing.
	Exclude bool `yaml:"exclude,omitempty"`
}

// NewHeader parses the header from the given data.
func NewHeader(data []byte) (header Header, err error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))

	dec.KnownFields(true)

	if err := dec.Decode(&header); err != nil {
		if errors.Is(err, io.EOF) {
			return Header{}, nil
		}

		return header, fmt.Errorf("parsing alias file: %w", err)
	}

	return header, nil
}
