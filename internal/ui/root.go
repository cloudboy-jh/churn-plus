package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// RootModel is the top-level model that switches between menu and TUI
type RootModel struct {
	mode            AppMode
	menuModel       MenuModel
	tuiModel        Model
	factory         *engine.Factory
	orchestrator    *engine.PipelineOrchestrator
	files           []*engine.FileInfo
	fileTree        *engine.FileNode
	cfg             *config.Config
	ctx             *engine.ProjectContext
	pipelineStarted bool
	initError       error
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
	factory *engine.Factory,
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
		factory:         factory,
		orchestrator:    nil, // Created on-demand
		files:           files,
		fileTree:        fileTree,
		cfg:             cfg,
		ctx:             ctx,
		pipelineStarted: autoRun,
		initError:       nil,
	}
}

// Init initializes the root model
func (rm RootModel) Init() tea.Cmd {
	if rm.mode == ModeMenu {
		return rm.menuModel.Init()
	}

	// Auto-run mode: initialize provider and start pipeline immediately
	if rm.pipelineStarted && rm.orchestrator == nil {
		provider, err := rm.factory.CreateProvider()
		if err != nil {
			rm.initError = err
			rm.mode = ModeMenu
			return rm.menuModel.Init()
		}

		orchestrator, err := rm.factory.CreateDefaultPipeline(provider)
		if err != nil {
			rm.initError = err
			rm.mode = ModeMenu
			return rm.menuModel.Init()
		}
		orchestrator.SetContext(rm.ctx)
		rm.orchestrator = orchestrator
		rm.tuiModel = NewModel(rm.fileTree, rm.orchestrator.GetPipeline(), rm.orchestrator, rm.files, rm.ctx, true)
		return tea.Batch(
			rm.tuiModel.Init(),
			rm.startPipeline(),
		)
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
		// Initialize provider and orchestrator now
		if rm.orchestrator == nil {
			provider, err := rm.factory.CreateProvider()
			if err != nil {
				rm.initError = err
				return rm, nil
			}

			orchestrator, err := rm.factory.CreateDefaultPipeline(provider)
			if err != nil {
				rm.initError = err
				return rm, nil
			}
			orchestrator.SetContext(rm.ctx)
			rm.orchestrator = orchestrator
		}

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
		view := rm.menuModel.View()
		// Show initialization error if any
		if rm.initError != nil {
			view += "\n\n" + theme.ErrorStyle.Render("⚠ Error: "+rm.initError.Error())
			view += "\n" + theme.MutedStyle.Render("Please configure API keys in Settings or via environment variables:")
			view += "\n" + theme.MutedStyle.Render("  • Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GOOGLE_API_KEY")
			view += "\n" + theme.MutedStyle.Render("  • Or configure in ~/.churn/config.json")
		}
		return view
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

// GetOrchestrator returns the orchestrator if it has been created
func (rm RootModel) GetOrchestrator() *engine.PipelineOrchestrator {
	return rm.orchestrator
}

// PipelineStartedMsg signals the pipeline has started
type PipelineStartedMsg struct {
	events <-chan engine.PipelineEvent
}
