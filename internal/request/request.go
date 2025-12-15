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
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

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

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: Buffer could get overrun.
	// A header that exceeds 1k would do that.
	buf := make([]byte, 1024)
	bufIdx := 0
	for !request.done() {
		n, err := reader.Read(buf[bufIdx:])
		// TODO: Decide what to do with error.
		if err != nil {
			return nil, err
		}

		bufIdx += n
		readN, err := request.parse(buf[:bufIdx+n])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}

	return request, nil

}
