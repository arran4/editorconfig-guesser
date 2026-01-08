package main

import (
	"flag"
	"fmt"
	"os"

	"editorconfig-guesser/internal/cli"
)

type Cmd interface {
	Execute(args []string) error
	Usage()
}

type InternalCommand struct {
	Exec      func(args []string) error
	UsageFunc func()
}

func (c *InternalCommand) Execute(args []string) error {
	return c.Exec(args)
}

func (c *InternalCommand) Usage() {
	c.UsageFunc()
}

type UserError struct {
	Err error
	Msg string
}

func (e *UserError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func NewUserError(err error, msg string) *UserError {
	return &UserError{Err: err, Msg: msg}
}

type RootCmd struct {
	*flag.FlagSet
	Commands    map[string]Cmd
	Version     string
	Commit      string
	Date        string
	saveFlag    bool
	verboseFlag bool
	args        []string
}

func (c *RootCmd) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	c.PrintDefaults()
	fmt.Fprintln(os.Stderr, "  Commands:")
	for name := range c.Commands {
		fmt.Fprintf(os.Stderr, "    %s\n", name)
	}
}

func (c *RootCmd) UsageRecursive() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	c.PrintDefaults()
	fmt.Fprintln(os.Stderr, "  Commands:")
}

func NewRoot(name, version, commit, date string) (*RootCmd, error) {
	c := &RootCmd{
		FlagSet:  flag.NewFlagSet(name, flag.ExitOnError),
		Commands: make(map[string]Cmd),
		Version:  version,
		Commit:   commit,
		Date:     date,
	}
	c.FlagSet.Usage = c.Usage

	c.BoolVar(&c.saveFlag, "s", false, "Save the file as .editorconfig")
	c.BoolVar(&c.saveFlag, "save", false, "Save the file as .editorconfig")

	c.BoolVar(&c.verboseFlag, "v", false, "Logs more than what is required")
	c.BoolVar(&c.verboseFlag, "verbose", false, "Logs more than what is required")
	c.Commands["help"] = &InternalCommand{
		Exec: func(args []string) error {
			for _, arg := range args {
				if arg == "-deep" {
					c.UsageRecursive()
					return nil
				}
			}
			c.Usage()
			return nil
		},
		UsageFunc: c.Usage,
	}
	c.Commands["usage"] = &InternalCommand{
		Exec: func(args []string) error {
			for _, arg := range args {
				if arg == "-deep" {
					c.UsageRecursive()
					return nil
				}
			}
			c.Usage()
			return nil
		},
		UsageFunc: c.Usage,
	}
	c.Commands["version"] = &InternalCommand{
		Exec: func(args []string) error {
			fmt.Printf("Version: %s\nCommit: %s\nDate: %s\n", c.Version, c.Commit, c.Date)
			return nil
		},
		UsageFunc: func() {
			fmt.Fprintf(os.Stderr, "Usage: %s version\n", os.Args[0])
		},
	}
	return c, nil
}

func (c *RootCmd) Execute(args []string) error {
	if err := c.Parse(args); err != nil {
		return NewUserError(err, fmt.Sprintf("flag parse error %s", err.Error()))
	}
	remainingArgs := c.Args()
	if len(remainingArgs) > 0 {
		if cmd, ok := c.Commands[remainingArgs[0]]; ok {
			return cmd.Execute(remainingArgs[1:])
		}
	}
	// Handle vararg args
	{
		varArgStart := 0
		if varArgStart > len(remainingArgs) {
			varArgStart = len(remainingArgs)
		}
		varArgs := remainingArgs[varArgStart:]
		c.args = varArgs
	}

	cli.Run(c.saveFlag, c.verboseFlag, c.args...)
	return nil
}
