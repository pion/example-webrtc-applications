// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// examples hosts the HTML/JS for different examples
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Examples represents the examples loaded from examples.json.
type Examples []*Example

// Example represents an example loaded from examples.json.
type Example struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsJS        bool
}

func main() {
	addr := flag.String("address", ":80", "Address to host the HTTP server on.")
	flag.Parse()

	log.Println("Listening on", *addr)
	err := serve(*addr)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func serve(addr string) error {
	// Load the examples
	examples, err := getExamples()
	if err != nil {
		return err
	}

	// Load the templates
	homeTemplate := template.Must(template.ParseFiles("index.html"))

	// Serve the required pages
	// DIY 'mux' to avoid additional dependencies
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		// Split up the URL. Expected parts:
		// 1: Base url
		// 2: "example"
		// 3: Example type: js
		// 4: Example folder, e.g.: data-channels
		// 5: Static file as part of the example
		parts := strings.Split(url, "/")
		if len(parts) > 4 &&
			parts[1] == "example" {
			exampleType := parts[2]
			exampleLink := parts[3]
			for _, example := range *examples {
				if example.Link != exampleLink {
					continue
				}
				fiddle := filepath.Join(exampleLink, "jsfiddle")
				if len(parts[4]) != 0 {
					http.StripPrefix("/example/"+exampleType+"/"+exampleLink+"/", http.FileServer(http.Dir(fiddle))).ServeHTTP(w, r)
					return
				}

				temp := template.Must(template.ParseFiles("example.html"))
				_, err = temp.ParseFiles(filepath.Join(fiddle, "demo.html"))
				if err != nil {
					panic(err)
				}

				data := struct {
					*Example
					JS bool
				}{
					example,
					exampleType == "js",
				}

				err = temp.Execute(w, data)
				if err != nil {
					panic(err)
				}
				return
			}
		}

		// Serve the main page
		err = homeTemplate.Execute(w, examples)
		if err != nil {
			panic(err)
		}
	})

	// Start the server
	return http.ListenAndServe(addr, nil) // nolint:gosec
}

// getExamples loads the examples from the examples.json file.
func getExamples() (*Examples, error) {
	file, err := os.Open("./examples.json")
	if err != nil {
		return nil, fmt.Errorf("failed to list examples (please run in the examples folder): %v", err) //nolint
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()

	var examples Examples
	err = json.NewDecoder(file).Decode(&examples)
	if err != nil {
		return nil, fmt.Errorf("failed to parse examples: %v", err) //nolint
	}

	for _, example := range examples {
		fiddle := filepath.Join(example.Link, "jsfiddle")
		js := filepath.Join(fiddle, "demo.js")
		if _, err := os.Stat(js); !os.IsNotExist(err) {
			example.IsJS = true
		}
	}

	return &examples, nil
}
