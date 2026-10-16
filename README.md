# Baltimore City Intelligence Engine (BCIE)

Go terminal capstone analyzing Baltimore public BPD arrests and 311 requests using concurrent CSV ingestion and geographic grid-cell matching.

## Requirements
- macOS, Go 1.22 or newer (https://go.dev/dl/)
- Internet access for real-data mode only
- No external Go dependencies

## Run
```sh
chmod +x run.sh
./run.sh                   # bundled synthetic demonstration records
./run.sh --real            # downloads two public CSV datasets (may be large)
go build -o bcie ./cmd/bcie
./bcie -arrests data/arrests_sample.csv -requests data/requests_sample.csv
go test ./...
go test -race ./...
```

Real-data URLs:
- BPD arrests: https://data.baltimorecity.gov/api/download/v1/items/619ec10c14b346f784a5a07bad4c43cd/csv?layers=0
- 311 2026: https://data.baltimorecity.gov/api/download/v1/items/de0ddaef68624e32a84e5197c5ac1829/csv?layers=0

If the ArcGIS export endpoint changes, manually download both datasets as CSV from their public item pages and run:
```sh
./bcie -arrests /path/to/arrests.csv -requests /path/to/311.csv
```

## Architecture
- `cmd/bcie/main.go`: CLI, parallel dataset loading, reporting
- `internal/engine/engine.go`: CSV ingestion, bounded worker queues, normalization, grid-cell matching, aggregations, terminal visualization
- `internal/engine/engine_test.go`: deterministic pipeline test
- `data/*_sample.csv`: synthetic fixtures (not actual Baltimore records)
- `reports/summary.txt`: generated terminal report

Each dataset is processed concurrently; within each CSV, a single producer streams rows to a bounded channel and worker goroutines normalize rows. Worker results are collected into a separate channel. Aggregation occurs after workers finish, avoiding shared mutable map writes. Memory is proportional to retained normalized records; downloads are capped at 150 MiB each. Go's garbage collector manages heap memory. Temporary downloaded CSVs are removed after use.

## Five stakeholder analyses
1. Highest combined arrest and 311 record volumes by grid cell
2. Infrastructure requests in grid cells that also contain arrests
3. Combined-volume resource-attention ranking (descriptive, not a validated priority score)
4. Arrest concentrations by grid cell
5. Unresolved 311 request concentrations by grid cell

## Interpretation and limitations
- Cells are 0.01-degree latitude/longitude bins, **not official neighborhoods**. Their ground area varies by latitude.
- Counts are not adjusted for population, reporting propensity, or geographic area.
- Arrests are not equivalent to reported crimes. The sources may cover different periods; compare only overlapping dates when making time-based claims.
- An absent/unknown 311 status is not counted as open; the classification of non-closed statuses is approximate.
- Rows without usable latitude/longitude are excluded and counted as skipped.
- No causal inference is supported.
- The 2026 311 feed changes over time. Dataset schemas and public export URLs can change; verify downloads before submitting.
- Demo CSVs are synthetic and should never be presented as real municipal results.

## Verification checklist
1. `go version` reports Go 1.22+.
2. `go test ./...` passes.
3. `go test -race ./...` passes.
4. `./run.sh` produces five titled analyses and `reports/summary.txt`.
5. `./run.sh --real` downloads and parses two genuine datasets; inspect any export or schema errors.
6. Archive source, sample CSVs, tests, run script, and README; exclude large downloaded files and generated binaries.
