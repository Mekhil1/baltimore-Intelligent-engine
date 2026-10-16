package engine

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Record struct {
	Kind, Category, Status, Month, Cell string
}

type Summary struct {
	Cell                                    string
	Arrests, Requests, Open, Infrastructure int
	Categories                              map[string]int
	Months                                  map[string]int
}

type Result struct {
	Areas                      []Summary
	Arrests, Requests, Skipped int
}

type Source struct {
	Kind, Path, URL string
}

func Fetch(ctx context.Context, s Source) (string, error) {
	if s.Path != "" {
		return s.Path, nil
	}
	if s.URL == "" {
		return "", fmt.Errorf("no source configured for %s", s.Kind)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s download HTTP %d", s.Kind, resp.StatusCode)
	}

	const limit int64 = 150 << 20

	f, err := os.CreateTemp("", "bcie-*.csv")
	if err != nil {
		return "", err
	}

	defer f.Close()

	n, err := io.Copy(f, io.LimitReader(resp.Body, limit+1))
	if err != nil {
		os.Remove(f.Name())
		return "", err
	}
	if n > limit {
		os.Remove(f.Name())
		return "", errors.New("dataset exceeds 150 MiB download limit; use a smaller export")
	}

	return f.Name(), nil
}

// norm normalizes CSV column names.
func norm(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.ToLower(strings.TrimSpace(s)))
}

func field(row []string, cols map[string]int, names ...string) string {
	for _, name := range names {
		if i, ok := cols[norm(name)]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
	}
	return ""
}

// coords validates approximate Baltimore-area coordinates.
func coords(row []string, cols map[string]int) (float64, float64, bool) {
	a := field(row, cols, "latitude", "lat", "y")
	b := field(row, cols, "longitude", "lon", "long", "lng", "x")

	lat, e1 := strconv.ParseFloat(a, 64)
	lon, e2 := strconv.ParseFloat(b, 64)

	if e1 != nil || e2 != nil ||
		math.IsNaN(lat) || math.IsNaN(lon) ||
		math.IsInf(lat, 0) || math.IsInf(lon, 0) ||
		lat < 39.1 || lat > 39.5 ||
		lon < -77 || lon > -76.3 {
		return 0, 0, false
	}

	return lat, lon, true
}

func parseDate(s string) string {
	s = strings.TrimSpace(s)

	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		return s[:7]
	}

	layouts := []string{
		"01/02/2006",
		"01/02/2006 03:04:05 PM",
		"1/2/2006 3:04:05 PM",
		"2006/01/02",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01")
		}
	}

	return "Unknown"
}

func ParseFile(ctx context.Context, kind, path string, workers int) ([]Record, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = false

	head, err := reader.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("%s header: %w", kind, err)
	}

	cols := make(map[string]int)
	for i, h := range head {
		cols[norm(h)] = i
	}

	has := func(names ...string) bool {
		for _, name := range names {
			if _, ok := cols[norm(name)]; ok {
				return true
			}
		}
		return false
	}

	if !has("latitude", "lat", "y") ||
		!has("longitude", "lon", "long", "lng", "x") {
		return nil, 0, fmt.Errorf(
			"%s CSV requires latitude/longitude columns; got %v",
			kind, head,
		)
	}

	if workers < 1 {
		workers = 1
	}

	jobs := make(chan []string, workers*4)
	out := make(chan Record, workers*4)
	readDone := make(chan error, 1)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for row := range jobs {
				lat, lon, ok := coords(row, cols)
				if !ok {
					continue
				}

				cell := fmt.Sprintf(
					"%.2f,%.2f",
					math.Floor(lat*100)/100,
					math.Floor(lon*100)/100,
				)

				category := field(
					row, cols,
					"chargeDescription", "charge",
					"description", "incidentType",
					"srType", "serviceRequestType",
					"serviceName", "type", "requestType",
				)

				if category == "" {
					category = "Unspecified"
				}

				status := field(
					row, cols,
					"status", "srStatus", "requestStatus",
				)

				date := field(
					row, cols,
					"arrestDate", "createdDate",
					"createdDateTime", "created",
					"requestDate", "requestedDate",
					"srCreatedDate", "date",
				)

				record := Record{
					Kind:     kind,
					Cell:     cell,
					Category: category,
					Status:   status,
					Month:    parseDate(date),
				}

				select {
				case out <- record:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	go func() {
		defer close(jobs)

		count := 0

		for {
			row, err := reader.Read()

			if err == io.EOF {
				readDone <- nil
				return
			}
			if err != nil {
				readDone <- err
				return
			}

			count++

			select {
			case jobs <- row:
			case <-ctx.Done():
				readDone <- ctx.Err()
				return
			}
		}
	}()

	var records []Record

	for record := range out {
		records = append(records, record)
	}

	readErr := <-readDone

	if readErr != nil {
		return nil, 0, readErr
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	countFile, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer countFile.Close()

	counter := csv.NewReader(countFile)
	counter.FieldsPerRecord = -1

	if _, err := counter.Read(); err != nil {
		return nil, 0, err
	}

	total := 0
	for {
		_, err := counter.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		total++
	}

	return records, total - len(records), nil
}

func isOpenStatus(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))

	switch s {
	case "open", "new", "pending", "in progress",
		"in-progress", "assigned", "submitted",
		"reopened":
		return true
	default:
		return false
	}
}

func Analyze(records []Record, skipped int) Result {
	areas := make(map[string]*Summary)
	result := Result{Skipped: skipped}

	for _, rec := range records {
		area := areas[rec.Cell]

		if area == nil {
			area = &Summary{
				Cell:       rec.Cell,
				Categories: make(map[string]int),
				Months:     make(map[string]int),
			}
			areas[rec.Cell] = area
		}

		if rec.Kind == "arrests" {
			area.Arrests++
			result.Arrests++
		} else if rec.Kind == "requests" {
			area.Requests++
			result.Requests++
			area.Categories[rec.Category]++

			if isOpenStatus(rec.Status) {
				area.Open++
			}

			category := strings.ToLower(rec.Category)

			keywords := []string{
				"pothole", "street", "road",
				"light", "trash", "water",
				"sewer", "sidewalk", "dump",
			}

			for _, word := range keywords {
				if strings.Contains(category, word) {
					area.Infrastructure++
					break
				}
			}
		}

		area.Months[rec.Kind+":"+rec.Month]++
	}

	for _, area := range areas {
		result.Areas = append(result.Areas, *area)
	}

	sort.Slice(result.Areas, func(i, j int) bool {
		return result.Areas[i].Cell < result.Areas[j].Cell
	})

	return result
}

func Top(r Result, score func(Summary) int, n int) []Summary {
	areas := append([]Summary(nil), r.Areas...)

	sort.Slice(areas, func(i, j int) bool {
		x := score(areas[i])
		y := score(areas[j])

		if x == y {
			return areas[i].Cell < areas[j].Cell
		}
		return x > y
	})

	if n < 0 {
		n = 0
	}
	if n < len(areas) {
		areas = areas[:n]
	}

	return areas
}

func Bar(n, max int) string {
	if max <= 0 || n <= 0 {
		return ""
	}

	k := int(float64(n) * 30 / float64(max))

	if k == 0 {
		k = 1
	}
	if k > 30 {
		k = 30
	}

	return strings.Repeat("#", k)
}

func ReportSection(r Result, option int) string {
	var b strings.Builder

	type analysis struct {
		title string
		score func(Summary) int
		keep  func(Summary) bool
	}

	analyses := map[int]analysis{
		1: {
			title: "Q1: Highest Arrest Activity",
			score: func(a Summary) int {
				return a.Arrests
			},
			keep: func(a Summary) bool {
				return a.Arrests > 0
			},
		},
		2: {
			title: "Q2: Highest 311 Request Volume",
			score: func(a Summary) int {
				return a.Requests
			},
			keep: func(a Summary) bool {
				return a.Requests > 0
			},
		},
		3: {
			title: "Q3: Arrest and Infrastructure Overlap",
			score: func(a Summary) int {
				return a.Infrastructure
			},
			keep: func(a Summary) bool {
				return a.Arrests > 0 && a.Infrastructure > 0
			},
		},
		4: {
			title: "Q4: Combined Municipal Activity",
			score: func(a Summary) int {
				return a.Arrests + a.Requests
			},
			keep: func(a Summary) bool {
				return a.Arrests > 0 && a.Requests > 0
			},
		},
		5: {
			title: "Q5: Unresolved 311 Requests",
			score: func(a Summary) int {
				return a.Open
			},
			keep: func(a Summary) bool {
				return a.Open > 0
			},
		},
	}

	selected, ok := analyses[option]
	if !ok {
		return "Invalid analysis selection.\n"
	}

	fmt.Fprintf(&b, "\n%s\n", selected.title)
	fmt.Fprintln(&b, strings.Repeat("-", 65))

	var filtered []Summary

	for _, area := range r.Areas {
		if selected.keep(area) {
			filtered = append(filtered, area)
		}
	}

	if len(filtered) == 0 {
		fmt.Fprintln(&b, "No matching records found.")
		return b.String()
	}

	sort.Slice(filtered, func(i, j int) bool {
		x := selected.score(filtered[i])
		y := selected.score(filtered[j])

		if x == y {
			return filtered[i].Cell < filtered[j].Cell
		}
		return x > y
	})

	if len(filtered) > 8 {
		filtered = filtered[:8]
	}

	max := selected.score(filtered[0])

	for _, area := range filtered {
		value := selected.score(area)

		fmt.Fprintf(
			&b,
			"%-13s %6d %-30s\n",
			area.Cell,
			value,
			Bar(value, max),
		)

		fmt.Fprintf(
			&b,
			"  Arrests: %d | 311: %d | Infrastructure: %d | Open: %d\n",
			area.Arrests,
			area.Requests,
			area.Infrastructure,
			area.Open,
		)
	}

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Grid cells are approximate, not official neighborhoods.")
	fmt.Fprintln(&b, "Counts are not population-normalized.")
	fmt.Fprintln(&b, "Arrests do not represent all crime incidents.")
	fmt.Fprintln(&b, "No causal relationship is implied.")

	return b.String()
}

// Report generates the complete five-question report.
func Report(r Result) string {
	var b strings.Builder

	fmt.Fprintln(&b, "B-MORE INSIGHT")
	fmt.Fprintln(&b, "Baltimore City Intelligence Report")
	fmt.Fprintln(&b, strings.Repeat("=", 65))

	fmt.Fprintf(
		&b,
		"Arrests: %d | 311 Requests: %d\n",
		r.Arrests,
		r.Requests,
	)

	fmt.Fprintf(
		&b,
		"Skipped Records: %d | Grid Cells: %d\n",
		r.Skipped,
		len(r.Areas),
	)

	for i := 1; i <= 5; i++ {
		b.WriteString(ReportSection(r, i))
	}

	return b.String()
}
