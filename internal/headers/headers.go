package headers

import (
	"bytes"
	"fmt"
)

// Headers is a map type that stores HTTP header key-value pairs
// Key: header name (string)
// Value: header value (string)
// Example: "Content-Type" -> "application/json"
type Headers map[string]string

var rn = []byte("\r\n")

// Constructor function to create empty instance of Headers
func NewHeaders() Headers {
	return map[string]string{}
}

// parseHeader parses a single header line (name: value format) into name and value strings.
// Input (fieldLine []byte): raw header line bytes
// Returns: (header name, header value, error)
func parseHeader(fieldLine []byte) (string, string, error) {

	// Split on first colon only (value may contain colons)
	// Example: "Authorization: Bearer:token:123" → ["Authorization", " Bearer:token:123"]
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	// Must have exactly 2 parts (name and value)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	// Get the name and value
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	// Header name cannot have trailing whitespace
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	return string(name), string(value), nil
}

// Parse is a method on the Headers type that extracts HTTP headers from raw bytes.
// Receiver (h Headers): called as h.Parse(data)
// Input (data []byte): raw bytes containing header lines
// Returns: (bytes consumed, all headers parsed, error)
func (h Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	done := false

	for {
		// Find the next \r\n separator
		idx := bytes.Index(data[read:], rn)

		// No separator = incomplete header, wait for more data
		if idx == -1 {
			break
		}

		// Empty line (\r\n at position 0) = end of all headers
		if idx == 0 {
			read += len(rn)
			done = true
			break
		}

		// Parse the header line (extract name and value)
		name, value, err := parseHeader(data[:idx])
		if err != nil {
			return 0, false, err
		}

		// Track bytes consumed (header line + separator)
		read += idx + len(rn)

		// Store header in the map
		h[name] = value

		// Advance past the header we just processed
		data = data[idx+len(rn):]
	}

	return read, done, nil
}
