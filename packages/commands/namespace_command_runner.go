package commands

import (
	"go.uber.org/zap"
	"reflect"
)

type CommandRunner struct {
	commands []Command
	log      *zap.Logger
}

func NewCommandRunner(looger *zap.Logger) *CommandRunner {
	return &CommandRunner{
		commands: []Command{},
		log:      looger,
	}
}

func (cr *CommandRunner) AddCommand(c Command) {
	cr.commands = append(cr.commands, c)
}

func (cr *CommandRunner) Run() {
	summaries := make(map[string]string)
	for _, command := range cr.commands {

		commandName := reflect.TypeOf(command).String()

		cr.log.Info("Running command ", zap.String("command", commandName))
		if err := command.Execute(); err != nil {
			cr.log.Error("failed command execution", zap.Error(err))
		}
		summaries[commandName] = command.GetSummary()

	}

	cr.log.Info("Command summary", zap.Any("summaries", summaries))

}
