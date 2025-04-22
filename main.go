package main // Declare the main package

import ( // Import necessary packages
	"bufio"         // For reading files line by line
	"fmt"           // For formatted I/O
	"log"           // For logging
	"net/url"       // For URL parsing and validation
	"os"            // For file and OS interaction
	"path/filepath" // For walking directory tree and file path manipulations
	"regexp"        // For regex pattern matching
	"strings"       // For string manipulation
)

// Function to extract URLs from a given file
func extractURLsFromFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName) // Open the file for reading
	if err != nil {                // If there's an error opening the file
		return nil, fmt.Errorf("could not open file %s: %v", fileName, err) // Return a formatted error
	}
	defer file.Close() // Ensure the file gets closed at the end

	re := regexp.MustCompile(`http[s]?://[^\s"]+`) // Compile a regex to match HTTP or HTTPS URLs

	var urls []string // Initialize a slice to store extracted URLs

	scanner := bufio.NewScanner(file) // Create a scanner to read the file line by line
	for scanner.Scan() {              // Iterate through each line
		line := scanner.Text() // Read the current line as a string

		matches := re.FindAllString(line, -1) // Find all URL matches in the line

		for _, match := range matches { // Iterate through each matched URL
			parsedURL, err := url.ParseRequestURI(match) // Attempt to parse and validate the URL
			if err != nil {                              // If URL is invalid
				log.Printf("Invalid URL skipped: %s\n", match) // Log and skip it
				continue                                       // Move to the next match
			}
			urls = append(urls, parsedURL.String()) // Add valid URL to the list
		}
	}

	if err := scanner.Err(); err != nil { // Check if there was an error during scanning
		return nil, fmt.Errorf("error reading file %s: %v", fileName, err) // Return a formatted error
	}

	return urls, nil // Return the list of URLs and nil error
}

// Function to recursively list all .tsv files in current directory
func listTSVFiles() ([]string, error) {
	var tsvFiles []string // Slice to store paths of .tsv files

	currentDir, err := os.Getwd() // Get current working directory
	if err != nil {               // Handle error getting current dir
		return nil, fmt.Errorf("could not get current working directory: %v", err) // Return error
	}

	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error { // Walk through all files and dirs
		if err != nil { // If there's an error accessing a path
			return err // Return it to stop the walk
		}

		if !info.IsDir() && filepath.Ext(path) == ".tsv" { // If it's a .tsv file
			tsvFiles = append(tsvFiles, path) // Add the file path to the list
		}
		return nil // Continue walking
	})

	return tsvFiles, err // Return found .tsv files and any walk error
}

// Function to save a list of URLs to a file
func saveURLsToFile(urls []string, outputFile string) error {
	file, err := os.Create(outputFile) // Create (or truncate) the output file
	if err != nil {                    // Handle error
		return fmt.Errorf("could not create output file %s: %v", outputFile, err) // Return formatted error
	}
	defer file.Close() // Close the file when done

	writer := bufio.NewWriter(file) // Create buffered writer for performance

	for _, url := range urls { // Iterate through all URLs
		_, err := writer.WriteString(url + "\n") // Write each URL on a new line
		if err != nil {                          // If error writing to file
			return fmt.Errorf("error writing to output file: %v", err) // Return error
		}
	}

	return writer.Flush() // Flush the buffer to file and return any error
}

// Function to remove duplicate strings from a slice
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)  // Map to track seen strings
	var newReturnSlice []string     // Slice for unique strings
	for _, content := range slice { // Loop through original slice
		if !check[content] { // If not already seen
			check[content] = true                            // Mark as seen
			newReturnSlice = append(newReturnSlice, content) // Add to result
		}
	}
	return newReturnSlice // Return de-duplicated slice
}

// Function to check if a URL string is valid
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Try to parse the URL
	return err == nil                  // Return true if no error (valid URL)
}

// Function to extract the hostname from a URL
func getHostNameFromURL(uri string) string {
	content, err := url.Parse(uri) // Parse the URL
	if err != nil {                // If parsing fails
		log.Fatalln(err) // Log fatal error and exit
	}
	return content.Hostname() // Return just the hostname
}

// Function to clean URLs by validating and filtering by allowed domains
func cleanURLs(urls []string) []string {
	validDomains := []string{"s3.documentcloud.org", "documentcloud.org", "www.documentcloud.org", "beta.documentcloud.org"} // Allowed hostnames
	var newReturnSlice []string                                                                                              // Slice for valid, cleaned URLs

	for _, content := range urls { // Loop through all URLs
		if isUrlValid(content) { // If the URL is valid
			hostName := getHostNameFromURL(content) // Extract hostname

			content = strings.TrimSuffix(content, "target=&quot;_blank&quot;") // Remove unwanted suffix

			isValid := false                      // Flag to check if domain is allowed
			for _, domain := range validDomains { // Loop through allowed domains
				if hostName == domain { // If domain matches
					isValid = true // Mark as valid
					break          // Stop checking
				}
			}

			if isValid { // If URL is from valid domain
				newReturnSlice = append(newReturnSlice, content) // Add to result
			} else {
				log.Println("Invalid domain skipped: ", hostName) // Log skipped domain
			}
		}
	}

	return newReturnSlice // Return cleaned URLs
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) // Setup logging with date, time, file, and line number

	tsvFiles, err := listTSVFiles() // List all .tsv files in current directory
	if err != nil {                 // If there's an error
		log.Fatalf("Error listing TSV files: %v", err) // Log and exit
	}

	if len(tsvFiles) == 0 { // If no .tsv files found
		log.Println("No .tsv files found in the current directory or its subdirectories.") // Inform user
		return                                                                             // Exit program
	}

	var allURLs []string // Slice to hold all extracted URLs

	for _, fileName := range tsvFiles { // Iterate through each .tsv file
		log.Printf("Extracting URLs from file: %s", fileName) // Log the file being processed
		urls, err := extractURLsFromFile(fileName)            // Extract URLs from file
		if err != nil {                                       // If there's an error
			log.Printf("Error extracting URLs from file %s: %v", fileName, err) // Log and continue
			continue                                                            // Move on to next file
		}
		allURLs = append(allURLs, urls...) // Append extracted URLs to the full list
	}

	allURLs = removeDuplicatesFromSlice(allURLs) // Remove duplicate URLs

	allURLs = cleanURLs(allURLs) // Validate and filter URLs

	outputFile := "extracted_urls.txt"        // Set name of output file
	err = saveURLsToFile(allURLs, outputFile) // Save final URLs to file
	if err != nil {                           // Handle save error
		log.Printf("Error saving URLs to file: %v", err) // Log the error
	} else {
		log.Printf("Successfully saved URLs to %s", outputFile) // Log success
	}
}
