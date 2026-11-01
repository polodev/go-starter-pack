package feature

import (
	"errors"
	"sync"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// QueueTestSuite runs the database queue against the real test database (MariaDB/MySQL by default).
type QueueTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}

var (
	handled      = make(chan string, 10)
	failingTries sync.Map
)

type recordingJob struct{}

func (j *recordingJob) Signature() string { return "test_recording_job" }
func (j *recordingJob) Handle(args ...any) error {
	handled <- args[0].(string)
	return nil
}

type failingJob struct{}

func (j *failingJob) Signature() string { return "test_failing_job" }
func (j *failingJob) Handle(args ...any) error {
	count, _ := failingTries.LoadOrStore("count", 0)
	failingTries.Store("count", count.(int)+1)
	return errors.New("the mail server said no")
}

func (s *QueueTestSuite) SetupTest() {
	s.RefreshDatabase()
	facades.Queue().Register([]contractsqueue.Job{&recordingJob{}, &failingJob{}})
}

func (s *QueueTestSuite) count(table, queue string) int64 {
	total, err := facades.DB().Table(table).Where("queue", queue).Count()
	s.Require().NoError(err)
	return total
}

func (s *QueueTestSuite) work(queue string, tries int, wait time.Duration) {
	worker := facades.Queue().Worker(contractsqueue.Args{Connection: "database", Queue: queue, Concurrent: 1, Tries: tries})
	go func() { _ = worker.Run() }()
	time.Sleep(wait)
	s.Require().NoError(worker.Shutdown())
}

func (s *QueueTestSuite) TestADispatchedJobWaitsInTheJobsTableUntilAWorkerRunsIt() {
	err := facades.Queue().Job(&recordingJob{}, []contractsqueue.Arg{{Type: "string", Value: "hello"}}).
		OnConnection("database").OnQueue("tests-ok").Dispatch()
	s.Require().NoError(err)
	s.Equal(int64(1), s.count("jobs", "tests-ok"), "stored, not run yet")
	s.Empty(handled)

	worker := facades.Queue().Worker(contractsqueue.Args{Connection: "database", Queue: "tests-ok", Concurrent: 1, Tries: 1})
	go func() { _ = worker.Run() }()
	select {
	case value := <-handled:
		s.Equal("hello", value, "the job got its argument")
	case <-time.After(15 * time.Second):
		s.Fail("the worker never ran the job")
	}
	s.Require().NoError(worker.Shutdown())
	s.Equal(int64(0), s.count("jobs", "tests-ok"), "removed after it ran")
}

func (s *QueueTestSuite) TestAWorkerOnlyTakesItsOwnQueueAndLeavesDelayedJobsWaiting() {
	s.Require().NoError(facades.Queue().Job(&recordingJob{}, []contractsqueue.Arg{{Type: "string", Value: "other"}}).
		OnConnection("database").OnQueue("tests-other").Dispatch())
	s.Require().NoError(facades.Queue().Job(&recordingJob{}, []contractsqueue.Arg{{Type: "string", Value: "later"}}).
		OnConnection("database").OnQueue("tests-delayed").Delay(time.Now().Add(time.Hour)).Dispatch())

	s.work("tests-delayed", 1, 3*time.Second)
	s.Empty(handled, "a delayed job waits, and another queue is not touched")
	s.Equal(int64(1), s.count("jobs", "tests-other"))
	s.Equal(int64(1), s.count("jobs", "tests-delayed"))
}

func (s *QueueTestSuite) TestAFailingJobIsTriedAgainThenMovesToFailedJobs() {
	failingTries.Store("count", 0)
	s.Require().NoError(facades.Queue().Job(&failingJob{}).OnConnection("database").OnQueue("tests-fail").Dispatch())

	s.work("tests-fail", 2, 5*time.Second)
	tries, _ := failingTries.Load("count")
	s.Equal(2, tries, "tried twice")
	s.Equal(int64(0), s.count("jobs", "tests-fail"), "no longer waiting")

	var failed []struct {
		UUID      string
		Exception string
	}
	s.Require().NoError(facades.DB().Table("failed_jobs").Where("queue", "tests-fail").Get(&failed))
	s.Require().Len(failed, 1, "one row in failed_jobs")
	s.Contains(failed[0].Exception, "the mail server said no")
	s.NotEmpty(failed[0].UUID)
}
