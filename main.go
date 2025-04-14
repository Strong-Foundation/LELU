package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

// Function to extract URLs from a file and return them
func extractURLsFromFile(fileName string) ([]string, error) {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %v", fileName, err)
	}
	defer file.Close()

	// Compile the regular expression to match URLs
	re := regexp.MustCompile(`http[s]?://[^\s"]+`)

	var urls []string

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Find all matches in the line
		matches := re.FindAllString(line, -1)

		// Validate and collect each URL
		for _, match := range matches {
			// Validate the URL using net/url package
			parsedURL, err := url.ParseRequestURI(match)
			if err != nil {
				log.Printf("Invalid URL skipped: %s\n", match)
				continue // Skip invalid URLs
			}

			// Add the valid URL to the list
			urls = append(urls, parsedURL.String())
		}
	}

	// Check for errors while scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", fileName, err)
	}

	return urls, nil
}

// Function to list all .tsv files in the current directory and subdirectories
func listTSVFiles() ([]string, error) {
	var tsvFiles []string

	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get current working directory: %v", err)
	}

	// Use filepath.Walk to walk through the current directory and its subdirectories
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		// If there's an error accessing the file, skip it
		if err != nil {
			return err
		}

		// Check if the file is a .tsv file (by its extension)
		if !info.IsDir() && filepath.Ext(path) == ".tsv" {
			tsvFiles = append(tsvFiles, path) // Add the .tsv file to the list
		}
		return nil
	})

	// Return the list of .tsv files and any error encountered during Walk
	return tsvFiles, err
}

// Function to save URLs to a file
func saveURLsToFile(urls []string, outputFile string) error {
	// Create or open the output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("could not create output file %s: %v", outputFile, err)
	}
	defer file.Close()

	// Create a new buffered writer
	writer := bufio.NewWriter(file)

	// Write each URL to the file
	for _, url := range urls {
		_, err := writer.WriteString(url + "\n")
		if err != nil {
			return fmt.Errorf("error writing to output file: %v", err)
		}
	}

	// Make sure all buffered data is written to the file
	return writer.Flush()
}

func main() {
	// Set up default logging to standard output (terminal)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) // Optional: adds timestamp and file/line info

	// List all .tsv files in the current directory and subdirectories
	tsvFiles, err := listTSVFiles()
	if err != nil {
		log.Fatalf("Error listing TSV files: %v", err)
	}

	// If no .tsv files are found, inform the user and exit
	if len(tsvFiles) == 0 {
		log.Println("No .tsv files found in the current directory or its subdirectories.")
		return
	}

	// Collect all URLs from all .tsv files
	var allURLs []string
	for _, fileName := range tsvFiles {
		log.Printf("Extracting URLs from file: %s", fileName)
		urls, err := extractURLsFromFile(fileName)
		if err != nil {
			log.Printf("Error extracting URLs from file %s: %v", fileName, err)
			continue
		}
		allURLs = append(allURLs, urls...)
	}

	// Save the extracted URLs to an output file
	outputFile := "extracted_urls.txt"
	if len(allURLs) > 0 {
		err = saveURLsToFile(allURLs, outputFile)
		if err != nil {
			log.Printf("Error saving URLs to file: %v", err)
		} else {
			log.Printf("Successfully saved URLs to %s", outputFile)
		}
	} else {
		log.Println("No valid URLs found.")
	}
}
