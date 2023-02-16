package main

import "github.com/urfave/cli/v2"

var (
	_ Commander = (*Command)(nil)
)

type (
	Commander interface {
		Init(name, usage, description string, aliases ...string) *cli.Command
		SetRunners(cli.BeforeFunc, cli.ActionFunc, cli.AfterFunc, cli.OnUsageErrorFunc)
		SetSubCommands(...*cli.Command)
		SetFlags(...cli.Flag)
		Get() *cli.Command
	}

	Command struct {
		C *cli.Command
	}

	CommandParams struct {
		Name         string
		Usage        string
		Description  string
		Aliases      []string
		Before       cli.BeforeFunc
		Action       cli.ActionFunc
		After        cli.AfterFunc
		OnUsageError cli.OnUsageErrorFunc
		Flags        []cli.Flag
		SubCommands  []*cli.Command
	}

	Option func(*Command)
)

func NewCommand(params *CommandParams, opts ...Option) *Command {
	c := &Command{}
	c.Init(params.Name, params.Usage, params.Description, params.Aliases...)
	c.SetRunners(params.Before, params.Action, params.After, params.OnUsageError)
	c.SetFlags(params.Flags...)
	c.SetSubCommands(params.SubCommands...)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (cmd *Command) Init(name, usage, description string, aliases ...string) *cli.Command {
	cmd.C = &cli.Command{
		Name:        name,
		Usage:       usage,
		Description: description,
		Aliases:     aliases,
	}

	return cmd.C
}

func (cmd *Command) SetRunners(before cli.BeforeFunc, action cli.ActionFunc, after cli.AfterFunc, onUsageError cli.OnUsageErrorFunc) {
	cmd.C.Before = before
	cmd.C.Action = action
	cmd.C.After = after
	cmd.C.OnUsageError = onUsageError
}

func (cmd *Command) SetSubCommands(subCommands ...*cli.Command) {
	cmd.C.Subcommands = subCommands
}

func (cmd *Command) SetFlags(flags ...cli.Flag) {
	cmd.C.Flags = flags
}

func (cmd *Command) Get() *cli.Command {
	return cmd.C
}
