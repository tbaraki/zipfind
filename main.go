package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip function extracts matching files from a zip file to the specified destination directory
func Unzip(src, dest, pattern string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Only extract files that match the pattern
		if !strings.Contains(f.Name, pattern) {
			continue
		}

		fpath := filepath.Join(dest, filepath.Base(f.Name))

		// Check for ZipSlip (directory traversal)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Skip directories
			continue
		}

		// Make File
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}

		// Check if the extracted file is a zip file and unzip it recursively
		if strings.HasSuffix(f.Name, ".zip") {
			if err := Unzip(fpath, dest, pattern); err != nil {
				return err
			}
			// Remove the nested zip file after extraction to avoid re-processing it
			os.Remove(fpath)
		}
	}
	return nil
}

func main() {
	// Define flags
	src := flag.String("src", "", "Source zip file")
	dest := flag.String("dest", "", "Destination directory")
	pattern := flag.String("pattern", "", "Pattern to match files")

	// Parse flags
	flag.Parse()

	// Check if src is provided
	if *src == "" {
		fmt.Println("Source zip file must be specified")
		flag.Usage()
		return
	}

	// Check if dest is provided
	if *dest == "" {
		fmt.Println("Destination directory must be specified")
		flag.Usage()
		return
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(*dest, os.ModePerm); err != nil {
		fmt.Println("Error creating destination directory:", err)
		return
	}

	// Call Unzip function
	err := Unzip(*src, *dest, *pattern)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Successfully unzipped to", *dest)
	}
}
