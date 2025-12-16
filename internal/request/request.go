package request

import (
	"bytes"
	"fmt"
	"io"
)

// We are trying to parse a line
// to get the following attributes
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// Method receives a pointer to a RequestLine struct and
// returns whether or not the HTTP Version is exactly 1.1
func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "HTTP/1.1"
}

// The general Request struct which contains
// RequestLine nested within (Method, HTTP Version, etc.) and
// the state of the request (init, done, error) to identify when to exit
type Request struct {
	RequestLine RequestLine
	state       parserState
}

// Initializes a new Request with StateInit and returns a pointer to it
func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

// Custom parserState type for different request states
type parserState string

const (
	StateInit  parserState = "init"
	StateDone  parserState = "done"
	StateError parserState = "error"
)

// Constants, including error codes and
// defined separator to indicate when to stop parsing
var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("ERRIR: Malformed Request Line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("ERROR: Unsupported HTTP Version")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("Request in error state.")
var SEPARATOR = []byte("\r\n")

func ParseRequestLine(b []byte) (*RequestLine, int, error) {
	// Search for the \r\n separator in the byte slice
	// Returns the index where it starts, or -1 if not found
	idx := bytes.Index(b, SEPARATOR)

	// If separator not found, we don't have a complete request line yet
	// Return nil (no data parsed), 0 bytes consumed, and nil error (not an error, just incomplete)
	// The caller will read more data and try again
	if idx == -1 {
		return nil, 0, nil
	}

	// Extract everything before the separator (the actual request line)
	// Example: "GET / HTTP/1.1\r\nHost: example.com" → startLine = "GET / HTTP/1.1"
	startLine := b[:idx]

	// Calculate how many bytes to skip to get past the separator
	// idx = position of \r, len(SEPARATOR) = 2 (\r\n), so we skip past both
	// Example: if idx=15, read=17 means skip to byte 17 (past the \r\n)
	read := idx + len(SEPARATOR)

	// Split the request line by spaces into parts
	// Should have exactly 3 parts: METHOD, REQUEST_TARGET, HTTP_VERSION
	// Example: "GET / HTTP/1.1" → ["GET", "/", "HTTP/1.1"]
	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	// Split the HTTP version part by "/" to validate format
	// Should be "HTTP" / "1.1"
	// Example: "HTTP/1.1" → ["HTTP", "1.1"]
	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	// Create the RequestLine struct with the parsed values
	// Convert byte slices to strings
	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	// Return the parsed RequestLine, bytes consumed (including \r\n), and no error
	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0

outer:
	for {
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			rl, n, err := ParseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateDone

		case StateDone:
			break outer
		}
		return read, nil
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

// RequestFromReader reads data from an io.Reader and parses it into a Request.
// It continuously reads data in chunks of up to 1024 bytes, parsing the HTTP request
// line until the request is complete (done) or an error occurs. The function maintains
// an internal buffer and shifts unconsumed data to the beginning of the buffer after
// each parse iteration. Returns a pointer to the parsed Request and any error encountered
// during reading or parsing.
func RequestFromReader(reader io.Reader) (*Request, error) {

	// Create a new request with StateInit
	request := newRequest()

	// Create a 1024 byte array to store the incoming info.
	// NOTE: Buffer could get overrun.
	buf := make([]byte, 1024)

	// Set the buffer index to the beginning
	bufIdx := 0

	// Loop until the request is complete or has an error
	for !request.done() {
		// Read up to 1024 bytes from TCP connection into buf starting at bufIdx
		// n is the number of bytes that were actually read
		n, err := reader.Read(buf[bufIdx:])
		// TODO: Decide what to do with error.
		if err != nil {
			return nil, err
		}

		// Advance buffer index by the number of bytes just read
		// bufIdx now represents total data currently in the buffer
		// Example: bufIdx was 0, read 256 bytes, now bufIdx = 256
		bufIdx += n

		// Parse the buffer to extract the HTTP request line
		// Returns readN = number of bytes consumed (including \r\n)
		// If readN is 0, there's incomplete data, loop continues to read more
		// If error, the request is malformed, return error
		readN, err := request.parse(buf[:bufIdx+n])
		if err != nil {
			return nil, err
		}

		// Shift unconsumed bytes to the front of the buffer
		// buf[readN:bufIdx] = all bytes after what was parsed
		// Example: if buffer has "GET / HTTP/1.1\r\nHost: example.com" and readN=18
		// This copies "Host: example.com" to the front
		copy(buf, buf[readN:bufIdx])

		// Adjust buffer index to account for consumed bytes
		// If bufIdx was 35 and readN was 18, bufIdx becomes 17
		// Now the unconsumed data occupies buf[0:17]
		bufIdx -= readN
	}

	return request, nil

}
