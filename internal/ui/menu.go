package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	orange = "\033[38;5;173m"
	tan    = "\033[38;5;180m"
	reset  = "\033[0m"
)

// Menu
func Menu(report string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Clear the terminal screen.
		fmt.Print("\033[H\033[2J")

		fmt.Println(orange)
		fmt.Println("============================================")
		fmt.Println("               B-MORE INSIGHT")
		fmt.Println("============================================")
		fmt.Print(reset)

		fmt.Println(tan + "       Baltimore City Data Analysis" + reset)
		fmt.Println("       Municipal Intelligence Engine")
		fmt.Println()

		fmt.Println(orange + "MAIN MENU" + reset)
		fmt.Println()
		fmt.Println("  1. Arrest Data Summary")
		fmt.Println("  2. 311 Service Requests")
		fmt.Println("  3. Compare Both Datasets")
		fmt.Println("  4. Resource Priority Analysis")
		fmt.Println("  5. Unresolved Service Requests")
		fmt.Println("  6. View Full Report")
		fmt.Println("  0. Exit")
		fmt.Println()

		fmt.Print(orange + "Select an option: " + reset)

		choice, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println("Unable to read selection:", err)
			return
		}

		choice = strings.TrimSpace(choice)

		if choice == "" && err == io.EOF {
			return
		}

		// Display the selected analysis.
		switch choice {
		case "1":
			showReport(
				"Arrest Data Summary",
				reportSection(report, 1),
			)

		case "2":
			showReport(
				"311 Service Requests",
				reportSection(report, 2),
			)

		case "3":
			showReport(
				"Cross-Dataset Comparison",
				reportSection(report, 3),
			)

		case "4":
			showReport(
				"Resource Priority Analysis",
				reportSection(report, 4),
			)

		case "5":
			showReport(
				"Unresolved Service Requests",
				reportSection(report, 5),
			)

		case "6":
			showReport(
				"Full Intelligence Report",
				report,
			)

		case "0":
			fmt.Println()
			fmt.Println(tan +
				"Thank you for using B-MORE INSIGHT!" +
				reset)
			return

		default:
			fmt.Println()
			fmt.Println("Invalid option. Please try again.")
		}

		if err == io.EOF {
			return
		}

		fmt.Print("\nPress ENTER to return to the menu...")
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
	}
}

func showReport(title, report string) {
	fmt.Println()

	fmt.Println(orange +
		"============================================" +
		reset)

	fmt.Println(tan + title + reset)

	fmt.Println(orange +
		"============================================" +
		reset)

	fmt.Println()
	fmt.Println(report)
}

// main reporter
func reportSection(report string, option int) string {
	if option < 1 || option > 5 {
		return "Invalid analysis selection."
	}

	lines := strings.Split(report, "\n")
	prefix := fmt.Sprintf("Q%d:", option)

	var selected []string
	capturing := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		//Detect func
		if strings.HasPrefix(trimmed, prefix) {
			capturing = true
		} else if isQuestionHeading(trimmed) {
			//fake end
			if capturing {
				break
			}
		}

		if capturing {
			selected = append(selected, line)
		}
	}

	if len(selected) == 0 {
		return "No analysis available for this selection."
	}

	return strings.TrimSpace(strings.Join(selected, "\n"))
}

func isQuestionHeading(line string) bool {
	return len(line) >= 3 &&
		line[0] == 'Q' &&
		line[1] >= '1' &&
		line[1] <= '5' &&
		line[2] == ':'
}
