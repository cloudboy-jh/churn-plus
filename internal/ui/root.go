package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
)

// RootModel is the top-level model that switches between menu and TUI
type RootModel struct {
	mode            AppMode
	menuModel       MenuModel
	tuiModel        Model
	orchestrator    *engine.PipelineOrchestrator
	files           []*engine.FileInfo
	fileTree        *engine.FileNode
	cfg             *config.Config
	ctx             *engine.ProjectContext
	pipelineStarted bool
}

// AppMode defines the current mode
type AppMode int

const (
	ModeMenu AppMode = iota
	ModeTUI
)

// NewRootModel creates the root model
func NewRootModel(
	cfg *config.Config,
	ctx *engine.ProjectContext,
	fileTree *engine.FileNode,
	orchestrator *engine.PipelineOrchestrator,
	files []*engine.FileInfo,
	autoRun bool,
) RootModel {
	mode := ModeMenu
	if autoRun {
		mode = ModeTUI
	}

	return RootModel{
		mode:            mode,
		menuModel:       NewMenuModel(cfg, ctx),
		orchestrator:    orchestrator,
		files:           files,
		fileTree:        fileTree,
		cfg:             cfg,
		ctx:             ctx,
		pipelineStarted: autoRun,
	}
}

// Init initializes the root model
func (rm RootModel) Init() tea.Cmd {
	if rm.mode == ModeMenu {
		return rm.menuModel.Init()
	}

	// Auto-run mode: start pipeline immediately
	if rm.pipelineStarted {
		return rm.startPipeline()
	}

	return nil
}

// Update handles messages
func (rm RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return rm, tea.Quit
		}

	case StartAnalysisMsg:
		// Switch from menu to TUI and start analysis
		rm.mode = ModeTUI
		rm.tuiModel = NewModel(rm.fileTree, rm.orchestrator.GetPipeline(), rm.orchestrator, rm.files, rm.ctx, true)
		rm.pipelineStarted = true
		return rm, tea.Batch(
			rm.tuiModel.Init(),
			rm.startPipeline(),
		)

	case PipelineStartedMsg:
		// Forward to TUI model
		if rm.mode == ModeTUI {
			var cmd tea.Cmd
			var model tea.Model
			model, cmd = rm.tuiModel.Update(msg)
			rm.tuiModel = model.(Model)
			return rm, cmd
		}
	}

	// Route to appropriate model
	if rm.mode == ModeMenu {
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = rm.menuModel.Update(msg)
		rm.menuModel = model.(MenuModel)
		return rm, cmd
	} else {
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = rm.tuiModel.Update(msg)
		rm.tuiModel = model.(Model)
		return rm, cmd
	}
}

// View renders the current mode
func (rm RootModel) View() string {
	if rm.mode == ModeMenu {
		return rm.menuModel.View()
	}
	return rm.tuiModel.View()
}

// startPipeline starts the analysis pipeline
func (rm RootModel) startPipeline() tea.Cmd {
	return func() tea.Msg {
		// Get events channel
		events := rm.orchestrator.Events()

		// Start pipeline in background
		go func() {
			if err := rm.orchestrator.Execute(context.Background(), rm.files); err != nil {
				// Error handling
			}
		}()

		return PipelineStartedMsg{events: events}
	}
}

// PipelineStartedMsg signals the pipeline has started
type PipelineStartedMsg struct {
	events <-chan engine.PipelineEvent
}
