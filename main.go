package main

import (
	"log"
	"os"
	"fmt"
	"bytes"
)

func main() {
	// Open the file named "messages.txt" and store the file handle in variable f
	// Also capture any error that might occur during the file open
	f, err := os.Open("messages.txt")
	
	// Check if an error occurred when opening the file
	if err != nil {
		// If yes, log the error and stop the program immediately
		log.Fatal("error", "error", err)
	}

	// Create an empty string variable to accumulate partial lines as we read chunks
	str := ""
	
	// Start an infinite loop that will read the entire file
	for {
		// Create a byte buffer (array) with space for exactly 8 bytes
		data := make([]byte, 8)
		
		// Read up to 8 bytes from the file into the buffer
		// n = how many bytes were actually read (might be less than 8)
		// err = any error that occurred (usually signals end of file)
		n, err := f.Read(data)
		
		// Check if an error occurred during the read (typically end of file)
		if err != nil {
			// If yes, break out of the loop and stop reading
			break
		}

		// Trim the data buffer to only include the bytes we actually read
		// This removes any garbage data in the unused positions
		data = data[:n]
		
		// Search for a newline character ('\n') in the data we just read
		// i = the position where newline was found, or -1 if not found
		if i := bytes.IndexByte(data, '\n'); i != -1 {
			// We found a newline! This means we have a complete line
			
			// Add everything BEFORE the newline to our accumulated string
			str += string(data[:i])
			
			// Remove everything up to and including the newline from the data
			// This leaves any text that came after the newline
			data = data[i + 1:]
			
			// Print the complete line we've assembled
			fmt.Printf("read: %s\n", str)
			
			// Reset the string accumulator to empty, ready for the next line
			str = ""
		}
		
		// Add any remaining bytes from this chunk to our string accumulator
		// (This could be the start of a new line if no newline was found)
		str += string(data)
	}

	// After the loop ends, check if there are any leftover bytes
	if len(str) != 0 {
		// If yes, this is a partial line at the end of the file (no newline after it)
		// Print it so we don't lose data
		fmt.Printf("read %s\n", str)
	}
}