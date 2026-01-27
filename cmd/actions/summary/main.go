package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

// JUnit XML types for parsing test results
type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string         `xml:"name,attr"`
	Classname string         `xml:"classname,attr"`
	Time      float64        `xml:"time,attr"`
	Failures  []junitFailure `xml:"failure"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

type ClassCoverageInfo struct {
	ClassName      string  `json:"className"`
	UncoveredLines []int   `json:"uncoveredLines"`
	TotalLines     int     `json:"totalLines"`
	CoveredCount   int     `json:"coveredCount"`
	UncoveredCount int     `json:"uncoveredCount"`
	Percentage     float64 `json:"percentage"`
	TopLevel       bool    `json:"topLevel,omitempty"`
	TopLevelClass  string  `json:"topLevelClass,omitempty"`
}

type CoverageSummary struct {
	Classes         []ClassCoverageInfo `json:"classes"`
	OverallCoverage float64             `json:"overallCoverage"`
	TotalLines      int                 `json:"totalLines"`
	CoveredLines    int                 `json:"coveredLines"`
	UncoveredLines  int                 `json:"uncoveredLines"`
}

type TestResults struct {
	Suite    junitTestSuite
	Coverage CoverageSummary
}

func main() {
	junitFile := flag.String("junit", "", "JUnit XML file with test results")
	coverageFile := flag.String("coverage", "", "JSON file with coverage data")
	flag.Parse()

	if *junitFile == "" && *coverageFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: summary --junit <results.xml> [--coverage <coverage.json>]\n")
		os.Exit(1)
	}

	var results TestResults

	if *junitFile != "" {
		suite, err := readJUnitXML(*junitFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading JUnit results: %v\n", err)
			os.Exit(1)
		}
		results.Suite = suite
	}

	if *coverageFile != "" {
		cov, err := readCoverageJSON(*coverageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading coverage data: %v\n", err)
			os.Exit(1)
		}
		results.Coverage = cov
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

func readJUnitXML(filename string) (junitTestSuite, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return junitTestSuite{}, err
	}
	var suite junitTestSuite
	if err := xml.Unmarshal(data, &suite); err != nil {
		return junitTestSuite{}, err
	}
	return suite, nil
}

func readCoverageJSON(filename string) (CoverageSummary, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return CoverageSummary{}, err
	}
	var cov CoverageSummary
	if err := json.Unmarshal(data, &cov); err != nil {
		return CoverageSummary{}, err
	}
	return cov, nil
}

func generateSummary(results *TestResults) string {
	var sb strings.Builder

	suite := results.Suite

	// Header with emoji and overall status
	allPassed := suite.Failures == 0
	statusEmoji := "✅"
	statusText := "All Tests Passed"
	if !allPassed {
		statusEmoji = "❌"
		statusText = "Some Tests Failed"
	}

	sb.WriteString(fmt.Sprintf("# %s Apex Test Results: %s\n\n", statusEmoji, statusText))

	// Test Summary Statistics
	if suite.Tests > 0 {
		passed := suite.Tests - suite.Failures

		sb.WriteString("## 📊 Test Summary\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Tests | **%d** |\n", suite.Tests))
		sb.WriteString(fmt.Sprintf("| ✅ Passed | **%d** |\n", passed))
		sb.WriteString(fmt.Sprintf("| ❌ Failed | **%d** |\n", suite.Failures))
		sb.WriteString(fmt.Sprintf("| ⏱️ Duration | **%s** |\n", formatDurationSeconds(suite.Time)))

		// Coverage Summary (inline in test summary table)
		if results.Coverage.TotalLines > 0 {
			coverage := results.Coverage.OverallCoverage
			coverageEmoji := getCoverageEmoji(coverage)

			sb.WriteString(fmt.Sprintf("| %s Code Coverage | **%.2f%%** |\n", coverageEmoji, coverage))
			sb.WriteString(fmt.Sprintf("| Lines Covered | **%d** / **%d** |\n",
				results.Coverage.CoveredLines, results.Coverage.TotalLines))
		}

		sb.WriteString("\n")
	}

	// Coverage visualization
	if results.Coverage.TotalLines > 0 {
		sb.WriteString("## 📈 Coverage Overview\n\n")
		coverage := results.Coverage.OverallCoverage
		barChart := generateCoverageBar(coverage)
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", barChart))

		// Coverage by class (aggregate inner classes into their top-level owner)
		topLevelClasses := aggregateCoverageByTopLevel(results.Coverage.Classes)
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
	if suite.Failures > 0 {
		sb.WriteString("## ❌ Failed Tests\n\n")
		for _, tc := range suite.TestCases {
			if len(tc.Failures) > 0 {
				sb.WriteString(fmt.Sprintf("### %s.%s\n\n", tc.Classname, tc.Name))
				for _, f := range tc.Failures {
					msg := f.Message
					if msg == "" {
						msg = f.Body
					}
					if msg != "" {
						sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", msg))
					}
				}
			}
		}
	}

	// Test timing details
	if len(suite.TestCases) > 0 {
		sb.WriteString("## ⏱️ Test Performance\n\n")

		// Slowest tests
		sortedByDuration := make([]junitTestCase, len(suite.TestCases))
		copy(sortedByDuration, suite.TestCases)
		sort.Slice(sortedByDuration, func(i, j int) bool {
			return sortedByDuration[i].Time > sortedByDuration[j].Time
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
			tc := sortedByDuration[i]
			statusEmoji := "✅"
			if len(tc.Failures) > 0 {
				statusEmoji = "❌"
			}
			sb.WriteString(fmt.Sprintf("| %s `%s.%s` | %s |\n",
				statusEmoji, tc.Classname, tc.Name, formatDurationSeconds(tc.Time)))
		}

		sb.WriteString("\n</details>\n\n")
	}

	// All tests (collapsible)
	if len(suite.TestCases) > 0 {
		sb.WriteString("## 📋 All Tests\n\n")
		sb.WriteString("<details>\n")
		sb.WriteString(fmt.Sprintf("<summary>View all %d tests</summary>\n\n", len(suite.TestCases)))
		sb.WriteString("| Status | Test | Duration |\n")
		sb.WriteString("|--------|------|----------|\n")

		for _, tc := range suite.TestCases {
			statusEmoji := "✅"
			if len(tc.Failures) > 0 {
				statusEmoji = "❌"
			}
			sb.WriteString(fmt.Sprintf("| %s | `%s.%s` | %s |\n",
				statusEmoji, tc.Classname, tc.Name, formatDurationSeconds(tc.Time)))
		}

		sb.WriteString("\n</details>\n\n")
	}

	return sb.String()
}

func formatDurationSeconds(seconds float64) string {
	ms := seconds * 1000
	if ms < 1000 {
		return fmt.Sprintf("%.0fms", ms)
	} else if ms < 60000 {
		return fmt.Sprintf("%.2fs", seconds)
	} else {
		minutes := int(seconds / 60)
		secs := seconds - float64(minutes*60)
		return fmt.Sprintf("%dm %.1fs", minutes, secs)
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

func aggregateCoverageByTopLevel(classes []ClassCoverageInfo) []ClassCoverageInfo {
	if len(classes) == 0 {
		return nil
	}

	agg := make(map[string]*ClassCoverageInfo)
	for _, cls := range classes {
		topName := cls.TopLevelClass
		if topName == "" {
			if strings.Contains(cls.ClassName, ".") {
				topName = cls.ClassName[:strings.Index(cls.ClassName, ".")]
			} else {
				topName = cls.ClassName
			}
		}

		entry := agg[topName]
		if entry == nil {
			entry = &ClassCoverageInfo{ClassName: topName}
			agg[topName] = entry
		}
		entry.TotalLines += cls.TotalLines
		entry.CoveredCount += cls.CoveredCount
	}

	result := make([]ClassCoverageInfo, 0, len(agg))
	for _, entry := range agg {
		if entry.TotalLines > 0 {
			if entry.CoveredCount > entry.TotalLines {
				entry.CoveredCount = entry.TotalLines
			}
			entry.UncoveredCount = entry.TotalLines - entry.CoveredCount
			entry.Percentage = float64(entry.CoveredCount) / float64(entry.TotalLines) * 100.0
		}
		result = append(result, *entry)
	}
	return result
}
