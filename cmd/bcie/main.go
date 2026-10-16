package main

import (
	"baltimore-intelligence-engine/internal/engine"
	"baltimore-intelligence-engine/internal/ui"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	arrestsURL  = "https://data.baltimorecity.gov/api/download/v1/items/619ec10c14b346f784a5a07bad4c43cd/csv?layers=0"
	requestsURL = "https://data.baltimorecity.gov/api/download/v1/items/de0ddaef68624e32a84e5197c5ac1829/csv?layers=0"
)

type outcome struct {
	records []engine.Record
	skipped int
	err     error
}

func main() {
	arrests := flag.String("arrests", "", "local arrests CSV")
	requests := flag.String("requests", "", "local 311 CSV")
	download := flag.Bool("download", false, "download real public data")
	workers := flag.Int("workers", runtime.NumCPU(), "workers per dataset")
	report := flag.String("report", "reports/summary.txt", "output report path")
	flag.Parse()

	if *download && (*arrests != "" || *requests != "") {
		log.Fatal("choose -download OR local -arrests and -requests")
	}

	if !*download && (*arrests == "" || *requests == "") {
		log.Fatal("provide both -arrests and -requests, or use -download")
	}

	if *workers < 1 {
		log.Fatal("workers must be >= 1")
	}

	ctx := context.Background()

	sources := []engine.Source{
		{Kind: "arrests", Path: *arrests},
		{Kind: "requests", Path: *requests},
	}

	if *download {
		sources[0].URL = arrestsURL
		sources[1].URL = requestsURL
	}

	fmt.Println("Loading Baltimore City datasets...")

	results := make([]outcome, len(sources))
	var wg sync.WaitGroup

	// Process both datasets concurrently.
	for i, source := range sources {
		wg.Add(1)

		go func(i int, source engine.Source) {
			defer wg.Done()

			path, err := engine.Fetch(ctx, source)
			if err != nil {
				results[i].err = err
				return
			}

			if source.Path == "" {
				defer os.Remove(path)
			}

			records, skipped, err := engine.ParseFile(
				ctx,
				source.Kind,
				path,
				*workers,
			)

			results[i] = outcome{
				records: records,
				skipped: skipped,
				err:     err,
			}
		}(i, source)
	}

	wg.Wait()

	for i, result := range results {
		if result.err != nil {
			log.Fatalf("%s: %v", sources[i].Kind, result.err)
		}
	}

	var all []engine.Record
	skipped := 0

	for _, result := range results {
		all = append(all, result.records...)
		skipped += result.skipped
	}

	if len(all) == 0 {
		log.Fatal("no valid geocoded records; check CSV fields")
	}

	// Analyze both datasets.
	analysis := engine.Analyze(all, skipped)
	output := engine.Report(analysis)

	// Save the report.
	if err := os.MkdirAll(filepath.Dir(*report), 0755); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(*report, []byte(output), 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Data processing complete.")
	fmt.Println("Report saved:", *report)

	// Launch
	ui.Menu(output)
}
