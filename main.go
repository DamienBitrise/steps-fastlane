package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/errorutil"
	. "github.com/bitrise-io/go-utils/v2/exitcode"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

func main() {
	exitCode := run()
	os.Exit(int(exitCode))
}

func run() ExitCode {
	logger := log.NewLogger()
	buildStep := createStep(logger)

	config, err := buildStep.ProcessConfig()
	if err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to process Step inputs: %w", err)))
		return Failure
	}

	dependenciesOpts := EnsureDependenciesOpts{
		GemVersions:    config.GemVersions,
		UseBundler:     config.GemVersions.fastlane.Found,
		WorkDir:        config.WorkDir,
		UpdateFastlane: config.UpdateFastlane,
	}

	if err := buildStep.InstallDependencies(dependenciesOpts); err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to install Step dependencies: %w", err)))
		return Failure
	}

	runOpts := createRunOptions(config)
	if err := buildStep.Run(runOpts); err != nil {
		buildStep.logger.Println()
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to execute Step: %w", err)))
		return Failure
	}

	buildStep.tracker.wait()

	return Success
}

func createStep(logger log.Logger) FastlaneRunner {
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepository)
	cmdFactory := command.NewFactory(envRepository)
	cmdLocator := env.NewCommandLocator()
	rbyFactory, err := ruby.NewCommandFactory(cmdFactory, cmdLocator)
	if err != nil {
		logger.Warnf("%s", err)
	}

	pathModifier := pathutil.NewPathModifier()
	tracker := newStepTracker(envRepository, logger)

	return NewFastlaneRunner(inputParser, logger, cmdLocator, cmdFactory, rbyFactory, pathModifier, tracker)
}

// FastlaneRunner ...
type FastlaneRunner struct {
	inputParser  stepconf.InputParser
	logger       log.Logger
	cmdFactory   command.Factory
	cmdLocator   env.CommandLocator
	rbyFactory   ruby.CommandFactory
	pathModifier pathutil.PathModifier
	tracker      stepTracker
}

// NewFastlaneRunner ...
func NewFastlaneRunner(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	commandLocator env.CommandLocator,
	cmdFactory command.Factory,
	rbyFactory ruby.CommandFactory,
	pathModifier pathutil.PathModifier,
	tracker stepTracker,
) FastlaneRunner {
	return FastlaneRunner{
		inputParser:  stepInputParser,
		logger:       logger,
		cmdLocator:   commandLocator,
		cmdFactory:   cmdFactory,
		rbyFactory:   rbyFactory,
		pathModifier: pathModifier,
		tracker:      tracker,
	}
}

func createRunOptions(config Config) RunOpts {
	return RunOpts{
		WorkDir:         config.WorkDir,
		AuthCredentials: config.AuthCredentials,
		LaneOptions:     config.LaneOptions,
		UseBundler:      config.GemVersions.fastlane.Found,
		GemVersions:     config.GemVersions,
		EnableCache:     config.EnableCache,
	}
}
