package jobs

type ExampleJob struct {
}

// Signature The name and signature of the job.
func (receiver *ExampleJob) Signature() string {
	return "example_job"
}

// Handle Execute the job.
func (receiver *ExampleJob) Handle(args ...any) error {
	return nil
}
