package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type TestMethodResult struct {
	TestName     string        `json:"testName"`
	ClassName    string        `json:"className"`
	MethodName   string        `json:"methodName"`
	Passed       bool          `json:"passed"`
	Duration     time.Duration `json:"duration"`
	DurationMs   float64       `json:"durationMs"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
}

type TestSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

type ClassCoverageInfo struct {
	ClassName      string  `json:"className"`
	CoveredLines   []int   `json:"coveredLines"`
	TotalLines     int     `json:"totalLines"`
	CoveredCount   int     `json:"coveredCount"`
	UncoveredCount int     `json:"uncoveredCount"`
	Percentage     float64 `json:"percentage"`
	TopLevel       bool    `json:"topLevel,omitempty"`
}

type CoverageSummary struct {
	Classes         []ClassCoverageInfo `json:"classes"`
	OverallCoverage float64             `json:"overallCoverage"`
	TotalLines      int                 `json:"totalLines"`
	CoveredLines    int                 `json:"coveredLines"`
	UncoveredLines  int                 `json:"uncoveredLines"`
}

type TestResults struct {
	Tests           []TestMethodResult `json:"tests"`
	Summary         TestSummary        `json:"summary"`
	Coverage        CoverageSummary    `json:"coverage"`
	StartTime       time.Time          `json:"startTime"`
	EndTime         time.Time          `json:"endTime"`
	TotalDuration   time.Duration      `json:"totalDuration"`
	TotalDurationMs float64            `json:"totalDurationMs"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: summary <results.json>\n")
		os.Exit(1)
	}

	resultsFile := os.Args[1]
	data, err := os.ReadFile(resultsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading results file: %v\n", err)
		os.Exit(1)
	}

	var results TestResults
	if err := json.Unmarshal(data, &results); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	summary := generateSummary(&results)

	// Write to GitHub Step Summary
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile != "" {
		f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening summary file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		if _, err := f.WriteString(summary); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing summary: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Generated GitHub Job Summary")
	} else {
		fmt.Print(summary)
	}
}

func generateSummary(results *TestResults) string {
	var sb strings.Builder

	// Header with emoji and overall status
	allPassed := results.Summary.Failed == 0
	statusEmoji := "✅"
	statusText := "All Tests Passed"
	if !allPassed {
		statusEmoji = "❌"
		statusText = "Some Tests Failed"
	}

	sb.WriteString(fmt.Sprintf("# %s Apex Test Results: %s\n\n", statusEmoji, statusText))

	// Test Summary Statistics
	sb.WriteString("## 📊 Test Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Tests | **%d** |\n", results.Summary.Total))
	sb.WriteString(fmt.Sprintf("| ✅ Passed | **%d** |\n", results.Summary.Passed))
	sb.WriteString(fmt.Sprintf("| ❌ Failed | **%d** |\n", results.Summary.Failed))
	sb.WriteString(fmt.Sprintf("| ⏱️ Duration | **%s** |\n", formatDuration(results.TotalDurationMs)))

	// Coverage Summary
	if results.Coverage.TotalLines > 0 {
		coverage := results.Coverage.OverallCoverage
		coverageEmoji := getCoverageEmoji(coverage)

		sb.WriteString(fmt.Sprintf("| %s Code Coverage | **%.2f%%** |\n", coverageEmoji, coverage))
		sb.WriteString(fmt.Sprintf("| Lines Covered | **%d** / **%d** |\n",
			results.Coverage.CoveredLines, results.Coverage.TotalLines))
	}

	sb.WriteString("\n")

	// Coverage visualization
	if results.Coverage.TotalLines > 0 {
		sb.WriteString("## 📈 Coverage Overview\n\n")
		coverage := results.Coverage.OverallCoverage
		barChart := generateCoverageBar(coverage)
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", barChart))

		// Coverage by class
		topLevelClasses := filterTopLevelClasses(results.Coverage.Classes)
		if len(topLevelClasses) > 0 {
			sb.WriteString("### Coverage by Class\n\n")
			sb.WriteString("<details>\n")
			sb.WriteString(fmt.Sprintf("<summary>View %d classes</summary>\n\n", len(topLevelClasses)))
			sb.WriteString("| Class | Coverage | Lines Covered |\n")
			sb.WriteString("|-------|----------|---------------|\n")

			// Sort classes by coverage percentage (descending)
			sortedClasses := make([]ClassCoverageInfo, len(topLevelClasses))
			copy(sortedClasses, topLevelClasses)
			sort.Slice(sortedClasses, func(i, j int) bool {
				return sortedClasses[i].Percentage > sortedClasses[j].Percentage
			})

			for _, cls := range sortedClasses {
				emoji := getCoverageEmoji(cls.Percentage)
				bar := generateMiniBar(cls.Percentage)
				sb.WriteString(fmt.Sprintf("| `%s` | %s %.1f%% %s | %d / %d |\n",
					cls.ClassName, emoji, cls.Percentage, bar, cls.CoveredCount, cls.TotalLines))
			}

			sb.WriteString("\n</details>\n\n")
		}
	}

	// Failed tests details
	if results.Summary.Failed > 0 {
		sb.WriteString("## ❌ Failed Tests\n\n")
		for _, test := range results.Tests {
			if !test.Passed {
				sb.WriteString(fmt.Sprintf("### %s.%s\n\n", test.ClassName, test.MethodName))
				if test.ErrorMessage != "" {
					sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", test.ErrorMessage))
				}
			}
		}
	}

	// Test timing details
	if len(results.Tests) > 0 {
		sb.WriteString("## ⏱️ Test Performance\n\n")

		// Slowest tests
		sortedByDuration := make([]TestMethodResult, len(results.Tests))
		copy(sortedByDuration, results.Tests)
		sort.Slice(sortedByDuration, func(i, j int) bool {
			return sortedByDuration[i].DurationMs > sortedByDuration[j].DurationMs
		})

		maxSlowest := 10
		if len(sortedByDuration) < maxSlowest {
			maxSlowest = len(sortedByDuration)
		}

		sb.WriteString("<details>\n")
		sb.WriteString("<summary>Top 10 Slowest Tests</summary>\n\n")
		sb.WriteString("| Test | Duration |\n")
		sb.WriteString("|------|----------|\n")

		for i := 0; i < maxSlowest; i++ {
			test := sortedByDuration[i]
			statusEmoji := "✅"
			if !test.Passed {
				statusEmoji = "❌"
			}
			sb.WriteString(fmt.Sprintf("| %s `%s.%s` | %s |\n",
				statusEmoji, test.ClassName, test.MethodName, formatDuration(test.DurationMs)))
		}

		sb.WriteString("\n</details>\n\n")
	}

	// All tests (collapsible)
	if len(results.Tests) > 0 {
		sb.WriteString("## 📋 All Tests\n\n")
		sb.WriteString("<details>\n")
		sb.WriteString(fmt.Sprintf("<summary>View all %d tests</summary>\n\n", len(results.Tests)))
		sb.WriteString("| Status | Test | Duration |\n")
		sb.WriteString("|--------|------|----------|\n")

		for _, test := range results.Tests {
			statusEmoji := "✅"
			if !test.Passed {
				statusEmoji = "❌"
			}
			sb.WriteString(fmt.Sprintf("| %s | `%s.%s` | %s |\n",
				statusEmoji, test.ClassName, test.MethodName, formatDuration(test.DurationMs)))
		}

		sb.WriteString("\n</details>\n\n")
	}

	return sb.String()
}

func formatDuration(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.0fms", ms)
	} else if ms < 60000 {
		return fmt.Sprintf("%.2fs", ms/1000)
	} else {
		minutes := int(ms / 60000)
		seconds := (ms - float64(minutes*60000)) / 1000
		return fmt.Sprintf("%dm %.1fs", minutes, seconds)
	}
}

func getCoverageEmoji(percentage float64) string {
	if percentage >= 80 {
		return "🟢"
	} else if percentage >= 60 {
		return "🟡"
	} else if percentage >= 40 {
		return "🟠"
	}
	return "🔴"
}

func generateCoverageBar(percentage float64) string {
	barLength := 50
	filled := int(math.Round((percentage / 100) * float64(barLength)))
	if filled < 0 {
		filled = 0
	} else if filled > barLength {
		filled = barLength
	}
	empty := barLength - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("Coverage: %.2f%% [%s]", percentage, bar)
}

func generateMiniBar(percentage float64) string {
	barLength := 10
	filled := int(math.Round((percentage / 100) * float64(barLength)))
	if filled < 0 {
		filled = 0
	} else if filled > barLength {
		filled = barLength
	}
	empty := barLength - filled

	return fmt.Sprintf("`%s%s`", strings.Repeat("█", filled), strings.Repeat("░", empty))
}

func filterTopLevelClasses(classes []ClassCoverageInfo) []ClassCoverageInfo {
	hasExplicit := false
	for _, cls := range classes {
		if cls.TopLevel {
			hasExplicit = true
			break
		}
	}

	var filtered []ClassCoverageInfo
	for _, cls := range classes {
		if hasExplicit {
			if !cls.TopLevel {
				continue
			}
		} else if strings.Contains(cls.ClassName, ".") {
			continue
		}
		filtered = append(filtered, cls)
	}
	return filtered
}
