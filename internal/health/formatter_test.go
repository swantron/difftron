package health

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHealthReport_ToJSON(t *testing.T) {
	report := &HealthReport{
		OverallCoverage: 85.5,
		ChangedCoverage: 90.0,
		TotalFiles:      10,
		ChangedFiles:    2,
		FileHealth:      make(map[string]*FileHealth),
		Insights:        []Insight{},
		Recommendations: []Recommendation{},
	}

	jsonData, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("ToJSON() returned empty data")
	}

	// Verify it's valid JSON
	var formatted FormatHealthReport
	if err := json.Unmarshal(jsonData, &formatted); err != nil {
		t.Errorf("ToJSON() returned invalid JSON: %v", err)
	}

	if formatted.Summary.OverallCoverage != 85.5 {
		t.Errorf("expected OverallCoverage 85.5, got %f", formatted.Summary.OverallCoverage)
	}
}

func TestHealthReport_ToMarkdown(t *testing.T) {
	report := &HealthReport{
		OverallCoverage: 85.5,
		ChangedCoverage: 90.0,
		TotalFiles:      10,
		ChangedFiles:    2,
		HealthyFiles:    8,
		AtRiskFiles:     1,
		RegressingFiles: 1,
		FileHealth:      make(map[string]*FileHealth),
		Insights:        []Insight{},
		Recommendations: []Recommendation{},
	}

	markdown := report.ToMarkdown()

	if !strings.Contains(markdown, "# Testing Health Report") {
		t.Error("markdown should contain header")
	}
	if !strings.Contains(markdown, "85.5") {
		t.Error("markdown should contain overall coverage")
	}
	if !strings.Contains(markdown, "90.0") {
		t.Error("markdown should contain changed coverage")
	}
}

func TestHealthReport_ToStructuredText(t *testing.T) {
	report := &HealthReport{
		OverallCoverage: 85.5,
		ChangedCoverage: 90.0,
		TotalFiles:      10,
		ChangedFiles:    2,
		FileHealth:      make(map[string]*FileHealth),
		Insights:        []Insight{},
		Recommendations: []Recommendation{},
	}

	text := report.ToStructuredText()

	if !strings.Contains(text, "TESTING HEALTH REPORT") {
		t.Error("text should contain report title")
	}
	if !strings.Contains(text, "85.5") {
		t.Error("text should contain overall coverage")
	}
}

func TestHealthReport_WithFileHealth(t *testing.T) {
	report := &HealthReport{
		OverallCoverage: 80.0,
		ChangedCoverage: 75.0,
		TotalFiles:      1,
		ChangedFiles:    1,
		FileHealth: map[string]*FileHealth{
			"test.go": {
				FilePath:              "test.go",
				CoveragePercentage:    80.0,
				ChangedCoveragePercentage: 75.0,
				IsNewFile:             false,
			},
		},
		Insights:        []Insight{},
		Recommendations: []Recommendation{},
	}

	markdown := report.ToMarkdown()
	if !strings.Contains(markdown, "test.go") {
		t.Error("markdown should contain file path")
	}

	jsonData, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	var formatted FormatHealthReport
	if err := json.Unmarshal(jsonData, &formatted); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(formatted.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(formatted.Files))
	}
	if formatted.Files[0].FilePath != "test.go" {
		t.Errorf("expected file path 'test.go', got %s", formatted.Files[0].FilePath)
	}
}
