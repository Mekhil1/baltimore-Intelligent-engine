# B-MORE INSIGHT

### Baltimore City Municipal Intelligence Engine

**Built with Go | Public Data Analysis | Concurrent Processing**

---

## Project Overview

The Go programming language was used to create the command-line data analysis software B-MORE INSIGHT. The study looks both 311 service requests and publicly accessible Baltimore City arrest data to find trends in local government activities across different regions.

The objective is to provide a structured, interactive terminal interface that facilitates the exploration and comprehension of city data.

B-MORE INSIGHT prioritizes understandable text-based visualizations, simple comparisons, and concise summaries above complex dashboards.

Baltimore's identity serves as the inspiration for the application's unique orange-and-tan terminal interface.

## Project Objectives

B-MORE INSIGHT addresses five research questions:

1. Which geographic areas have the highest arrest activity?
2. Which areas generate the highest volume of 311 service requests?
3. Where do arrest activity and infrastructure-related complaints overlap?
4. Which areas show the highest combined municipal activity?
5. Which areas have the most unresolved 311 service requests?

These analyses describe patterns in the available records and do not establish causal relationships.

## Key Features

- **Concurrent Data Processing:** Uses Go goroutines and worker pools to process CSV records.
- **Public Dataset Integration:** Supports Baltimore City arrest and 311 service request data.
- **Geographic Analysis:** Groups records into approximate geographic grid cells.
- **Interactive Main Menu:** Allows users to select individual analyses.
- **Text-Based Visualizations:** Displays ranked results using ASCII bar charts.
- **Report Generation:** Saves a comprehensive analysis report.
- **Input Validation:** Handles invalid coordinates and tracks skipped records.
- **Custom Terminal Design:** Uses warm orange-and-tan colors for a clean interface.

## Technologies Used

| Technology          | Purpose                                     |
| ------------------- | ------------------------------------------- |
| Go                  | Core programming language                   |
| Goroutines          | Concurrent execution                        |
| Channels            | Worker communication                        |
| Go standard library | CSV parsing, HTTP requests, file operations |
| Git and GitHub      | Version control and project hosting         |
| VS Code             | Development environment                     |

## Project Structure

```text
baltimore-intelligence-engine/
├── cmd/
│   └── bcie/
│       └── main.go
├── data/
│   ├── arrests.csv
│   └── requests.csv
├── internal/
│   ├── engine/
│   │   ├── engine.go
│   │   └── engine_test.go
│   └── ui/
│       └── menu.go
├── reports/
│   └── summary.txt
├── .gitignore
├── go.mod
├── README.md
└── run.sh
```

*Dataset filenames above are the recommended names. Adjust them to match the actual files in the repository.*

## Getting Started

### Requirements

- Go installed on your computer
- Terminal or command-line environment
- Baltimore City arrest and 311 datasets in CSV format

### Installation

Clone the repository:

```bash
git clone https://github.com/Mekhil1/baltimore-Intelligent-engine.git
```

Navigate to the project directory:

```bash
cd baltimore-Intelligent-engine
```

### Running the Application

Run the application using locally stored CSV files:

```bash
go run ./cmd/bcie \
  -arrests data/arrests.csv \
  -requests data/requests.csv
```

The application processes both datasets, generates an analysis report, and opens the interactive terminal menu.

If your CSV files have different names, update the file paths accordingly.

## Interactive Main Menu

The application provides the following options:

```text
----------------------------------------
             B-MORE INSIGHT
----------------------------------------

      Baltimore City Data Analysis
      Municipal Intelligence Engine

MAIN MENU

  1. Arrest Data Summary
  2. 311 Service Requests
  3. Compare Both Datasets
  4. Resource Priority Analysis
  5. Unresolved Service Requests
  6. View Full Report
  0. Exit

Select an option:
```

Each analysis displays ranked geographic areas with corresponding counts and text-based visualizations.

## Data Sources

The project uses publicly available datasets from the Baltimore City Open Data portal.

**Baltimore City Arrest Data**

[https://data.baltimorecity.gov/api/download/v1/items/619ec10c14b346f784a5a07bad4c43cd/csv?layers=0](https://data.baltimorecity.gov/api/download/v1/items/619ec10c14b346f784a5a07bad4c43cd/csv?layers=0)

**Baltimore City 311 Service Requests**

[https://data.baltimorecity.gov/api/download/v1/items/de0ddaef68624e32a84e5197c5ac1829/csv?layers=0](https://data.baltimorecity.gov/api/download/v1/items/de0ddaef68624e32a84e5197c5ac1829/csv?layers=0)

The application supports local CSV files and includes functionality for downloading public datasets.

Large datasets may require smaller exports due to download-size limitations.

## Data Analysis Methodology

The application processes geographic coordinates and groups valid records into approximately 0.01-degree grid cells.

For each geographic cell, the engine calculates:

- Total arrest records
- Total 311 service requests
- Infrastructure-related service requests
- Unresolved service requests
- Monthly record counts

The results are ranked according to the selected research question.

**Important limitations:**

- Geographic grid cells are not official Baltimore neighborhoods.
- Counts are not normalized by population.
- Arrest records do not represent all crime incidents.
- Service request activity may reflect reporting behavior.
- Overlapping patterns do not demonstrate causation.

## Testing and Verification

The application uses Go's built-in testing and validation tools.

Run the following commands from the project root:

```bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

These checks were successfully executed during development.

## Project Purpose

B-MORE INSIGHT demonstrates how concurrent programming, data processing, and geographic aggregation can be combined to explore real-world municipal datasets.

The project emphasizes efficient processing, readable output, responsible interpretation of public data, and a user-friendly command-line experience.

## Developer

**Mekhi Lewis**

Computer Science Student
Morgan State University

**Project:** B-MORE INSIGHT — Baltimore City Municipal Intelligence Engine

---

*Built for Baltimore. Powered by data.*
