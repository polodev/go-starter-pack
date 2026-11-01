package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
	"goravel/app/modular"
)

func Jobs() []queue.Job {
	return append(modular.Collect(Modules(), func(m modular.HasJobs) []queue.Job { return m.Jobs() }),
		&jobs.ExampleJob{},
	)
}
