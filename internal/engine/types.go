package engine

import "time"

// Severity levels for findings
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Finding represents a single issue discovered during analysis
type Finding struct {
	File      string   `json:"file"`
	LineStart int      `json:"line_start"`
	LineEnd   int      `json:"line_end"`
	Severity  Severity `json:"severity"`
	Kind      string   `json:"kind"`      // e.g., "unreachable-code", "unused-import", "security"
	Message   string   `json:"message"`
	Pass      string   `json:"pass"`      // Which pass generated this finding
	Code      string   `json:"code,omitempty"` // Optional: code snippet
}

// ProjectContext holds metadata about the analyzed project
type ProjectContext struct {
	RootPath     string            `json:"root_path"`
	Languages    []string          `json:"languages"`
	Frameworks   []string          `json:"frameworks"`
	Tools        []string          `json:"tools"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	FileCount    int               `json:"file_count"`
}

// PassStatus represents the state of a pipeline pass
type PassStatus string

const (
	PassPending   PassStatus = "pending"
	PassRunning   PassStatus = "running"
	PassCompleted PassStatus = "completed"
	PassFailed    PassStatus = "failed"
)

// Pass represents a single analysis pass in the pipeline
type Pass struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      PassStatus `json:"status"`
	Model       string     `json:"model"`       // e.g., "claude-3.5-sonnet", "gpt-4-turbo"
	Provider    string     `json:"provider"`    // e.g., "anthropic", "openai"
	StartTime   time.Time  `json:"start_time,omitempty"`
	EndTime     time.Time  `json:"end_time,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// Pipeline represents the multi-pass analysis workflow
type Pipeline struct {
	Passes   []*Pass    `json:"passes"`
	Findings []*Finding `json:"findings"`
	Context  *ProjectContext `json:"context"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

// PipelineEvent represents events emitted during pipeline execution
type PipelineEvent struct {
	Type    PipelineEventType
	Pass    *Pass
	Finding *Finding
	Message string
	Error   error
}

type PipelineEventType string

const (
	EventPassStarted   PipelineEventType = "pass_started"
	EventPassProgress  PipelineEventType = "pass_progress"
	EventPassCompleted PipelineEventType = "pass_completed"
	EventPassFailed    PipelineEventType = "pass_failed"
	EventFindingAdded  PipelineEventType = "finding_added"
)

// FileInfo represents metadata about a single file
type FileInfo struct {
	Path     string
	Language string
	Size     int64
	Lines    int
}

// AnalysisReport is the final output structure
type AnalysisReport struct {
	Version     string          `json:"version"`
	Timestamp   time.Time       `json:"timestamp"`
	Context     *ProjectContext `json:"context"`
	Findings    []*Finding      `json:"findings"`
	Summary     ReportSummary   `json:"summary"`
	Pipeline    []*Pass         `json:"pipeline"`
}

// ReportSummary provides aggregate statistics
type ReportSummary struct {
	FilesAnalyzed int                 `json:"files_analyzed"`
	FindingCount  int                 `json:"finding_count"`
	BySeverity    map[Severity]int    `json:"by_severity"`
	ByKind        map[string]int      `json:"by_kind"`
	Duration      float64             `json:"duration_seconds"`
}
