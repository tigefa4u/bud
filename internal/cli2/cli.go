package cli

import (
	"context"
	"errors"
	"os"
	"runtime"

	"github.com/livebud/bud"
	"github.com/livebud/bud/internal/config"
	"github.com/livebud/bud/internal/pubsub"
	"github.com/livebud/bud/internal/sh"
	"github.com/livebud/bud/package/commander"
	"github.com/livebud/bud/package/socket"
)

func New() *CLI {
	return &CLI{
		&sh.Command{
			Dir:    ".",
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Env:    os.Environ(),
		},
		"info",
		pubsub.New(),
		nil,
		nil,
		nil,
	}
}

type CLI struct {
	*sh.Command
	Log string

	// Passed in for testing
	Bus          pubsub.Client
	WebListener  socket.Listener
	DevListener  socket.Listener
	FileListener socket.Listener
}

func (c *CLI) Parse(ctx context.Context, args ...string) error {
	// Check that we have a valid Go version
	if err := config.CheckGoVersion(runtime.Version()); err != nil {
		return err
	}

	in := new(custom)
	cmd := commander.New("bud").Writer(c.Stdout)
	cmd.Flag("chdir", "change the working directory").Short('C').String(&c.Dir).Default(c.Dir)
	cmd.Flag("help", "show this help message").Short('h').Bool(&in.Help).Default(false)
	cmd.Flag("log", "filter logs with this pattern").Short('L').String(&c.Log).Default("info")
	cmd.Args("args").Strings(&in.Args)
	cmd.Run(func(ctx context.Context) error { return c.runCustom(ctx, in) })

	{ // $ bud run
		in := new(bud.Run)
		cmd := cmd.Command("run", "run your app in dev")
		cmd.Flag("embed", "embed assets").Bool(&in.Embed).Default(false)
		cmd.Flag("hot", "hot reloading").Bool(&in.Hot).Default(true)
		cmd.Flag("minify", "minify assets").Bool(&in.Minify).Default(false)
		cmd.Flag("watch", "watch for changes").Bool(&in.Watch).Default(true)
		cmd.Flag("listen", "address to listen to").String(&in.WebAddress).Default(":3000")
		cmd.Flag("listen-dev", "dev address to listen to").String(&in.DevAddress).Default(":35729")
		cmd.Run(func(ctx context.Context) error { return c.Run(ctx, in) })
	}

	// Parse the arguments
	if err := cmd.Parse(ctx, args); err != nil {
		// Treat cancellation as a non-error
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	return nil
}