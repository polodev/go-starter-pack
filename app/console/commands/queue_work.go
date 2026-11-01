package commands

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
)

// QueueWork is `queue:work`, Laravel's worker: a separate process for Supervisor. (`go run .` also runs a worker
// inside the web server when the queue is not "sync"; both can run together, each job is taken by one worker.)
type QueueWork struct {
}

// Signature The name and signature of the console command.
func (r *QueueWork) Signature() string {
	return "queue:work"
}

// Description The console command description.
func (r *QueueWork) Description() string {
	return "Run jobs from a queue until stopped (Ctrl+C or SIGTERM)"
}

// Extend The console command extend.
func (r *QueueWork) Extend() command.Extend {
	return command.Extend{
		Category: "queue",
		Flags: []command.Flag{
			&command.StringFlag{Name: "connection", Usage: "queue connection (default: queue.default)"},
			&command.StringFlag{Name: "queue", Value: "default", Usage: "queue name"},
			&command.IntFlag{Name: "tries", Value: 3, Usage: "times a job is tried before it moves to failed_jobs"},
			&command.IntFlag{Name: "concurrent", Value: 1, Usage: "jobs run at the same time"},
		},
	}
}

// Handle Execute the console command.
func (r *QueueWork) Handle(ctx console.Context) error {
	connection := ctx.Option("connection")
	if connection == "" {
		connection = facades.Config().GetString("queue.default")
	}
	worker := facades.Queue().Worker(queue.Args{
		Connection: connection,
		Queue:      ctx.Option("queue"),
		Tries:      ctx.OptionInt("tries"),
		Concurrent: ctx.OptionInt("concurrent"),
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		_ = worker.Shutdown()
	}()

	ctx.Info("working " + connection + "/" + ctx.Option("queue") + " (Ctrl+C to stop)")
	return worker.Run()
}
