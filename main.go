package main

import (
	"log"
	"os"
	"fmt"
	"bytes"
	"io"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	
	// Create a channel with a buffer of 1, in string format
	out := make(chan string, 1)

	// Run anonymous function in background
	go func() {

		// Close file when function ends.
		defer f.Close()

		// Close the channel, ending background function.
		defer close(out)

		// Create empty string.
		str := ""

		// Start infinite loop.
		for {

			// Attempt to create an 8-byte buffer.
			data := make([]byte, 8)

			// (n, err) is a tuple
			// n = length, err = error code, if any
			n, err := f.Read(data)

			// If error, break loop, which will close channel
			if err != nil {
				break
			}

			// Get data up to length of 8-byte buffer.
			data = data[:n]
			
			// Search data for a '\n' new line.
			// IndexByte returns position of \n, -1 if not found.
			i := bytes.IndexByte(data, '\n')

		// If a new line is found (i is not -1)...
		if i != -1 {

			// Add everything before the newline to str
			str += string(data[:i])

			// Keep only the data after the newline for next iteration
			data = data[i + 1:]

				// The channel sends out whatever is in str.
				out <- str

				// String is cleared out.
				str = ""
			}

			// Append to string any remaining data.
			str += string(data)
		}

		// Handle case where file ends without a newline
		// Send any leftover data that didn't have a newline after it
		if len(str) != 0 {
			out <- str
		}
	} ()

	// Despite what happens inside loop, exit function.
	return out
}

func main() {

	// Open the file and store file handle in f, and error in err.
	f, err := os.Open("messages.txt")

	// If error, log it and quit.
	if err != nil {
		log.Fatal("error", "error", err)
	}

	// Create a background channel called lines
	lines := getLinesChannel(f)

	// Go through each line
	for line := range lines {

		// Print each line
		fmt.Printf("%s\n", line)
	}
}