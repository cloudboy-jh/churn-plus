package engine

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// FindingsAggregator collects and manages findings from multiple passes
type FindingsAggregator struct {
	findings []*Finding
	seen     map[string]bool // For deduplication
}

// NewFindingsAggregator creates a new findings aggregator
func NewFindingsAggregator() *FindingsAggregator {
	return &FindingsAggregator{
		findings: make([]*Finding, 0),
		seen:     make(map[string]bool),
	}
}

// Add adds a finding, deduplicating if necessary
func (fa *FindingsAggregator) Add(finding *Finding) {
	// Create a hash of the finding for deduplication
	hash := fa.hashFinding(finding)

	if fa.seen[hash] {
		// Already seen this finding, skip
		return
	}

	fa.seen[hash] = true
	fa.findings = append(fa.findings, finding)
}

// AddMultiple adds multiple findings
func (fa *FindingsAggregator) AddMultiple(findings []*Finding) {
	for _, f := range findings {
		fa.Add(f)
	}
}

// GetAll returns all findings
func (fa *FindingsAggregator) GetAll() []*Finding {
	return fa.findings
}

// GetBySeverity returns findings filtered by severity
func (fa *FindingsAggregator) GetBySeverity(severity Severity) []*Finding {
	result := make([]*Finding, 0)
	for _, f := range fa.findings {
		if f.Severity == severity {
			result = append(result, f)
		}
	}
	return result
}

// GetByFile returns findings for a specific file
func (fa *FindingsAggregator) GetByFile(file string) []*Finding {
	result := make([]*Finding, 0)
	for _, f := range fa.findings {
		if f.File == file {
			result = append(result, f)
		}
	}
	return result
}

// GetByKind returns findings of a specific kind
func (fa *FindingsAggregator) GetByKind(kind string) []*Finding {
	result := make([]*Finding, 0)
	for _, f := range fa.findings {
		if f.Kind == kind {
			result = append(result, f)
		}
	}
	return result
}

// Sort sorts findings by severity (high to low), then by file
func (fa *FindingsAggregator) Sort() {
	sort.Slice(fa.findings, func(i, j int) bool {
		// Sort by severity first
		severityOrder := map[Severity]int{
			SeverityCritical: 0,
			SeverityHigh:     1,
			SeverityMedium:   2,
			SeverityLow:      3,
		}

		if severityOrder[fa.findings[i].Severity] != severityOrder[fa.findings[j].Severity] {
			return severityOrder[fa.findings[i].Severity] < severityOrder[fa.findings[j].Severity]
		}

		// Then by file
		if fa.findings[i].File != fa.findings[j].File {
			return fa.findings[i].File < fa.findings[j].File
		}

		// Then by line number
		return fa.findings[i].LineStart < fa.findings[j].LineStart
	})
}

// Count returns the total number of findings
func (fa *FindingsAggregator) Count() int {
	return len(fa.findings)
}

// CountBySeverity returns counts grouped by severity
func (fa *FindingsAggregator) CountBySeverity() map[Severity]int {
	counts := map[Severity]int{
		SeverityLow:      0,
		SeverityMedium:   0,
		SeverityHigh:     0,
		SeverityCritical: 0,
	}

	for _, f := range fa.findings {
		counts[f.Severity]++
	}

	return counts
}

// CountByKind returns counts grouped by kind
func (fa *FindingsAggregator) CountByKind() map[string]int {
	counts := make(map[string]int)

	for _, f := range fa.findings {
		counts[f.Kind]++
	}

	return counts
}

// hashFinding creates a unique hash for deduplication
func (fa *FindingsAggregator) hashFinding(f *Finding) string {
	// Hash based on file, line, kind, and message
	data := fmt.Sprintf("%s:%d:%d:%s:%s", f.File, f.LineStart, f.LineEnd, f.Kind, f.Message)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// GenerateReport creates an analysis report from findings
func GenerateReport(
	ctx *ProjectContext,
	findings []*Finding,
	passes []*Pass,
	startTime time.Time,
	endTime time.Time,
) *AnalysisReport {
	aggregator := NewFindingsAggregator()
	aggregator.AddMultiple(findings)
	aggregator.Sort()

	duration := endTime.Sub(startTime).Seconds()

	summary := ReportSummary{
		FilesAnalyzed: ctx.FileCount,
		FindingCount:  aggregator.Count(),
		BySeverity:    aggregator.CountBySeverity(),
		ByKind:        aggregator.CountByKind(),
		Duration:      duration,
	}

	return &AnalysisReport{
		Version:   "0.1.0",
		Timestamp: time.Now(),
		Context:   ctx,
		Findings:  aggregator.GetAll(),
		Summary:   summary,
		Pipeline:  passes,
	}
}

// SaveReport saves a report to .churn/reports/
func SaveReport(projectRoot string, report *AnalysisReport) error {
	reportsDir := filepath.Join(projectRoot, ".churn", "reports")

	// Ensure directory exists
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("churn-report-%s.json", report.Timestamp.Format("2006-01-02T15-04-05"))
	path := filepath.Join(reportsDir, filename)

	// Marshal report to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// LoadReport loads a report from a file
func LoadReport(path string) (*AnalysisReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read report: %w", err)
	}

	var report AnalysisReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse report: %w", err)
	}

	return &report, nil
}

// ListReports returns all reports in the .churn/reports/ directory
func ListReports(projectRoot string) ([]string, error) {
	reportsDir := filepath.Join(projectRoot, ".churn", "reports")

	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read reports directory: %w", err)
	}

	reports := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			reports = append(reports, filepath.Join(reportsDir, entry.Name()))
		}
	}

	return reports, nil
}
