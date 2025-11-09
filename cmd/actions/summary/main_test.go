package main

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateSummaryIncludesCoverageDetails(t *testing.T) {
	results := &TestResults{
		Summary: TestSummary{
			Total:  3,
			Passed: 3,
			Failed: 0,
		},
		Coverage: CoverageSummary{
			OverallCoverage: 75.0,
			TotalLines:      40,
			CoveredLines:    30,
			UncoveredLines:  10,
			Classes: []ClassCoverageInfo{
				{
					ClassName:     "Alpha",
					CoveredCount:  20,
					TotalLines:    30,
					Percentage:    66.6,
					TopLevel:      true,
					TopLevelClass: "Alpha",
				},
				{
					ClassName:     "Beta",
					CoveredCount:  10,
					TotalLines:    20,
					Percentage:    50.0,
					TopLevel:      true,
					TopLevelClass: "Beta",
				},
				{
					ClassName:     "Alpha.InnerHelper",
					CoveredCount:  8,
					TotalLines:    10,
					Percentage:    80.0,
					TopLevel:      false,
					TopLevelClass: "Alpha",
				},
			},
		},
		Tests: []TestMethodResult{
			{TestName: "Alpha.testOne", ClassName: "Alpha", MethodName: "testOne", Passed: true, DurationMs: 100},
		},
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
	}

	summary := generateSummary(results)

	if !strings.Contains(summary, "🟡 Code Coverage | **75.00%**") {
		t.Fatalf("Summary missing coverage percentage: %s", summary)
	}
	if !strings.Contains(summary, "| Lines Covered | **30** / **40** |") {
		t.Fatalf("Summary missing line totals: %s", summary)
	}
	if !strings.Contains(summary, "Coverage: 75.00%") {
		t.Fatalf("Coverage bar missing or incorrect: %s", summary)
	}
	if !strings.Contains(summary, "`Alpha` | 🟡 70.0%") {
		t.Fatalf("Coverage by class table missing aggregated Alpha row: %s", summary)
	}
	if !strings.Contains(summary, "28 / 40") {
		t.Fatalf("Aggregated Alpha coverage counts missing: %s", summary)
	}
	if !strings.Contains(summary, "10 / 20") {
		t.Fatalf("Aggregated Beta coverage counts missing: %s", summary)
	}
	if strings.Contains(summary, "Alpha.InnerHelper") {
		t.Fatalf("Inner classes should be hidden from coverage table: %s", summary)
	}
}

func TestGenerateSummaryShowsFailedTests(t *testing.T) {
	results := &TestResults{
		Summary: TestSummary{
			Total:  2,
			Passed: 1,
			Failed: 1,
		},
		Coverage: CoverageSummary{},
		Tests: []TestMethodResult{
			{TestName: "Beta.testOne", ClassName: "Beta", MethodName: "testOne", Passed: false, DurationMs: 200, ErrorMessage: "boom"},
		},
	}

	summary := generateSummary(results)

	if !strings.Contains(summary, "Some Tests Failed") {
		t.Fatalf("Summary should mention failures: %s", summary)
	}
	if !strings.Contains(summary, "## ❌ Failed Tests") {
		t.Fatalf("Failed tests section missing: %s", summary)
	}
	if !strings.Contains(summary, "Beta.testOne") {
		t.Fatalf("Failed test details missing: %s", summary)
	}
}

func TestGenerateSummaryIncludesClassesWhenTopLevelUnknown(t *testing.T) {
	results := &TestResults{
		Summary: TestSummary{
			Total:  1,
			Passed: 1,
			Failed: 0,
		},
		Coverage: CoverageSummary{
			OverallCoverage: 10,
			TotalLines:      10,
			CoveredLines:    1,
			UncoveredLines:  9,
			Classes: []ClassCoverageInfo{
				{
					ClassName:    "Gamma",
					CoveredCount: 1,
					TotalLines:   10,
					Percentage:   10,
				},
				{
					ClassName:    "Delta.Inner",
					CoveredCount: 2,
					TotalLines:   4,
					Percentage:   50,
				},
			},
		},
	}

	summary := generateSummary(results)
	if !strings.Contains(summary, "`Gamma`") {
		t.Fatalf("Class Gamma should be included when top-level flag missing: %s", summary)
	}
	if !strings.Contains(summary, "`Delta`") {
		t.Fatalf("Dotted class names should roll up to their prefix when metadata missing: %s", summary)
	}
}
