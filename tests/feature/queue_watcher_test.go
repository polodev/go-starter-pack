package feature

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	queuewatcher "goravel/app-modules/queue-watcher"
	"goravel/app-modules/queue-watcher/models"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/tests"
)

// QueueWatcherTestSuite covers the queue monitor: every job run is recorded, and the screens show, retry and
// delete waiting and failed jobs for people with queue.view / queue.manage.
type QueueWatcherTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestQueueWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(QueueWatcherTestSuite))
}

type watchedJob struct{}

func (j *watchedJob) Signature() string        { return "test_watched_job" }
func (j *watchedJob) Handle(args ...any) error { return nil }

type watchedFailingJob struct{}

func (j *watchedFailingJob) Signature() string { return "test_watched_failing_job" }
func (j *watchedFailingJob) Handle(args ...any) error {
	return errors.New("the payment provider timed out")
}

func (s *QueueWatcherTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	auth.SendMail = func(auth.Outgoing) error { return nil }
	facades.Queue().Register([]contractsqueue.Job{&watchedJob{}, &watchedFailingJob{}})
	queuewatcher.TrackAll()
}

// must fails the test when a query builder call (result, error) fails.
func (s *QueueWatcherTestSuite) must(_ any, err error) { s.Require().NoError(err) }

func (s *QueueWatcherTestSuite) signedInAs(role string) *browser {
	email := role + "@example.com"
	user, _, err := auth.CreateUser("Person "+role, email, "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	if role != "customer" {
		s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	}
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	response, _ := b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
	response.AssertStatus(303)
	b.get("/") // the session is new after login: pick up its CSRF token
	return b
}

// run dispatches job on queue and lets a worker work until the history has want runs of it.
func (s *QueueWatcherTestSuite) run(job contractsqueue.Job, queueName string, tries int, want int64) {
	s.Require().NoError(facades.Queue().Job(job).OnConnection("database").OnQueue(queueName).Dispatch())
	worker := facades.Queue().Worker(contractsqueue.Args{Connection: "database", Queue: queueName, Concurrent: 1, Tries: tries})
	go func() { _ = worker.Run() }()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		count, _ := facades.Orm().Query().Model(&models.QueueJob{}).Where("job_class", job.Signature()).
			Where("status <> ?", models.StatusRunning).Count()
		pending, _ := facades.Orm().Query().Model(&models.PendingJob{}).Where("queue", queueName).Count()
		if count >= want && pending == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond) // let the worker write failed_jobs
	s.Require().NoError(worker.Shutdown())
}

func (s *QueueWatcherTestSuite) runs(signature string) []models.QueueJob {
	var runs []models.QueueJob
	s.Require().NoError(facades.Orm().Query().Where("job_class", signature).OrderBy("id").Find(&runs))
	return runs
}

func (s *QueueWatcherTestSuite) TestEveryRunOfEveryJobIsRecordedEvenGoravelsMailJob() {
	s.run(&watchedJob{}, "watch-ok", 1, 1)
	runs := s.runs("test_watched_job")
	s.Require().Len(runs, 1)
	s.Equal(models.StatusFinished, runs[0].Status)
	s.NotNil(runs[0].StartedAt)
	s.NotNil(runs[0].FinishedAt)
	s.Nil(runs[0].ExceptionMessage)

	s.run(&watchedFailingJob{}, "watch-fail", 2, 2)
	runs = s.runs("test_watched_failing_job")
	s.Require().Len(runs, 2, "each try is a run")
	for _, run := range runs {
		s.Equal(models.StatusFailed, run.Status)
		s.Require().NotNil(run.ExceptionMessage)
		s.Contains(*run.ExceptionMessage, "the payment provider timed out")
	}

	mailJob, err := facades.Queue().JobStorer().Get("goravel_send_mail_job")
	s.Require().NoError(err)
	s.Contains(fmt.Sprintf("%T", mailJob), "trackedJob", "Goravel's own mail job is tracked too")
	queuewatcher.TrackAll()
	again, _ := facades.Queue().JobStorer().Get("goravel_send_mail_job")
	s.Same(mailJob, again, "tracking twice does not wrap twice")
}

func (s *QueueWatcherTestSuite) TestOnlyPeopleWithQueueViewSeeTheMonitorAndOnlyQueueManageChangesIt() {
	for _, role := range []string{"customer", "new_employee"} {
		response, _ := s.signedInAs(role).get("/admin/queue-watcher")
		response.AssertStatus(403)
	}
	employee := s.signedInAs("employee")
	for _, path := range []string{"/admin/queue-watcher", "/admin/queue-watcher/jobs", "/admin/queue-watcher/pending", "/admin/queue-watcher/failed"} {
		response, body := employee.get(path)
		response.AssertStatus(200)
		s.NotContains(body, "data-testid=\"retry-all-button\"")
		s.NotContains(body, "data-testid=\"purge-button\"")
	}
	_, dashboard := employee.get("/admin")
	s.Contains(dashboard, `href="/admin/queue-watcher"`, "the sidebar links the monitor")

	s.run(&watchedFailingJob{}, "watch-fail", 1, 1)
	var failed models.FailedJob
	s.Require().NoError(facades.Orm().Query().Where("queue", "watch-fail").First(&failed))
	response, _ := employee.post("/admin/queue-watcher/failed/"+strconv.Itoa(int(failed.ID))+"/retry", url.Values{})
	response.AssertStatus(403)
	count, _ := facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(1), count, "nothing changed after the refusal")
}

func (s *QueueWatcherTestSuite) TestAFailedJobIsShownRetriedAndLoggedByAnAdmin() {
	s.run(&watchedFailingJob{}, "watch-fail", 1, 1)
	var failed models.FailedJob
	s.Require().NoError(facades.Orm().Query().Where("queue", "watch-fail").First(&failed))
	admin := s.signedInAs("admin")

	response, body := admin.get("/admin/queue-watcher/failed")
	response.AssertStatus(200)
	s.Contains(body, `data-testid="failed-row"`)
	s.Contains(body, "test_watched_failing_job")
	s.Contains(body, "the payment provider timed out")

	id := strconv.Itoa(int(failed.ID))
	response, body = admin.get("/admin/queue-watcher/failed/" + id)
	response.AssertStatus(200)
	s.Contains(body, `data-testid="failed-error"`)
	s.Contains(body, `data-testid="retry-button"`)
	s.Contains(body, `Dashboard`)

	response, _ = admin.postWithToken("/admin/queue-watcher/failed/"+id+"/retry", url.Values{}, "")
	response.AssertStatus(419)
	count, _ := facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(1), count, "without the CSRF token nothing happens")

	response, _ = admin.post("/admin/queue-watcher/failed/"+id+"/retry", url.Values{})
	s.Equal("/admin/queue-watcher/failed?saved=retried", location(response))
	count, _ = facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(0), count, "it left failed_jobs")
	waiting, _ := facades.Orm().Query().Model(&models.PendingJob{}).Where("queue", "watch-fail").Count()
	s.Equal(int64(1), waiting, "and waits on its queue again")

	logged, _ := facades.DB().Table("activity_log").Where("log_name", "queue").Where("event", "retried").Count()
	s.Equal(int64(1), logged)
}

func (s *QueueWatcherTestSuite) TestRetryAllDeleteAndFlush() {
	s.run(&watchedFailingJob{}, "watch-fail", 1, 1)
	s.run(&watchedFailingJob{}, "watch-fail-2", 1, 2)
	admin := s.signedInAs("admin")
	response, _ := admin.post("/admin/queue-watcher/failed/retry-all", url.Values{})
	s.Equal("/admin/queue-watcher/failed?saved=retried", location(response))
	failedCount, _ := facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(0), failedCount)
	pendingCount, _ := facades.Orm().Query().Model(&models.PendingJob{}).Count()
	s.Equal(int64(2), pendingCount)

	s.must(facades.DB().Table("jobs").Where("1 = 1").Delete())
	s.run(&watchedFailingJob{}, "watch-fail-3", 1, 3)
	s.run(&watchedFailingJob{}, "watch-fail-4", 1, 4)
	var first models.FailedJob
	s.Require().NoError(facades.Orm().Query().OrderBy("id").First(&first))
	response, _ = admin.post("/admin/queue-watcher/failed/"+strconv.Itoa(int(first.ID))+"/delete", url.Values{})
	s.Equal("/admin/queue-watcher/failed?saved=deleted", location(response))
	failedCount, _ = facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(1), failedCount)

	response, _ = admin.post("/admin/queue-watcher/failed/flush", url.Values{})
	s.Equal("/admin/queue-watcher/failed?saved=flushed", location(response))
	failedCount, _ = facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(0), failedCount)
}

func (s *QueueWatcherTestSuite) TestAFailedJobWithoutUuidIsNotRetriedAndNothingElseIs() {
	s.run(&watchedFailingJob{}, "watch-fail", 1, 1)
	s.must(facades.DB().Table("failed_jobs").Insert(map[string]any{
		"uuid": "", "connection": "database", "queue": "odd", "payload": "{}", "exception": "imported", "failed_at": time.Now().UTC(),
	}))
	var odd models.FailedJob
	s.Require().NoError(facades.Orm().Query().Where("queue", "odd").First(&odd))

	response, _ := s.signedInAs("admin").post("/admin/queue-watcher/failed/"+strconv.Itoa(int(odd.ID))+"/retry", url.Values{})
	s.Equal("/admin/queue-watcher/failed/"+strconv.Itoa(int(odd.ID))+"?error=retry", location(response))
	count, _ := facades.Orm().Query().Model(&models.FailedJob{}).Count()
	s.Equal(int64(2), count, "an empty uuid must not retry every failed job")
	pending, _ := facades.Orm().Query().Model(&models.PendingJob{}).Count()
	s.Equal(int64(0), pending)
}

func (s *QueueWatcherTestSuite) TestPendingJobsAreListedAndDeletableButMailArgumentsStayHidden() {
	s.Require().NoError(facades.Queue().Job(&watchedJob{}, []contractsqueue.Arg{{Type: "string", Value: "order-42"}}).
		OnConnection("database").OnQueue("watch-later").Delay(time.Now().Add(time.Hour)).Dispatch())
	now := time.Now().UTC()
	s.must(facades.DB().Table("jobs").Insert(map[string]any{
		"queue": "mail", "attempts": 0, "available_at": now, "created_at": now,
		"payload": `{"uuid":"u-1","signature":"goravel_send_mail_job","args":[{"type":"string","value":"https://example.com/reset-password/SECRET-TOKEN"}],"chain":[]}`,
	}))
	var delayed, mail models.PendingJob
	s.Require().NoError(facades.Orm().Query().Where("queue", "watch-later").First(&delayed))
	s.Require().NoError(facades.Orm().Query().Where("queue", "mail").First(&mail))

	admin := s.signedInAs("admin")
	response, body := admin.get("/admin/queue-watcher")
	response.AssertStatus(200)
	s.Contains(body, `data-testid="count-delayed">1<`)
	s.Contains(body, `data-testid="count-waiting">1<`)

	_, body = admin.get("/admin/queue-watcher/pending")
	s.Contains(body, "test_watched_job")
	s.Contains(body, "goravel_send_mail_job")

	_, body = admin.get("/admin/queue-watcher/pending/" + strconv.Itoa(int(delayed.ID)))
	s.Contains(body, `data-testid="show-state">delayed<`)
	s.Contains(body, "order-42", "ordinary arguments are shown")

	_, body = admin.get("/admin/queue-watcher/pending/" + strconv.Itoa(int(mail.ID)))
	s.NotContains(body, "SECRET-TOKEN", "a mail's links are never shown")
	s.Contains(body, "hidden")

	response, _ = admin.post("/admin/queue-watcher/pending/"+strconv.Itoa(int(delayed.ID))+"/delete", url.Values{})
	s.Equal("/admin/queue-watcher/pending?saved=deleted", location(response))
	left, _ := facades.Orm().Query().Model(&models.PendingJob{}).Count()
	s.Equal(int64(1), left)

	s.must(facades.DB().Table("jobs").Where("id", mail.ID).Update("reserved_at", now))
	response, _ = admin.post("/admin/queue-watcher/pending/"+strconv.Itoa(int(mail.ID))+"/delete", url.Values{})
	s.Contains(location(response), "error=running")
	left, _ = facades.Orm().Query().Model(&models.PendingJob{}).Count()
	s.Equal(int64(1), left, "a job a worker holds is not deleted")

	response, _ = admin.get("/admin/queue-watcher/pending/999999")
	response.AssertStatus(404)
}

func (s *QueueWatcherTestSuite) TestUnfinishedRunsBecomeStaleAndOldHistoryIsPurged() {
	old := time.Now().UTC().AddDate(0, 0, -40)
	twoHoursAgo := time.Now().UTC().Add(-2 * time.Hour)
	recent := time.Now().UTC()
	insert := func(jobID, status string, created, started time.Time) {
		s.must(facades.DB().Table("queue_jobs").Insert(map[string]any{
			"job_id": jobID, "job_class": "test_history", "status": status, "attempt": 1,
			"started_at": started, "created_at": created, "updated_at": created,
		}))
	}
	insert("old-finished", models.StatusFinished, old, old)
	insert("stuck", models.StatusRunning, twoHoursAgo, twoHoursAgo)
	insert("busy", models.StatusRunning, recent, recent)
	insert("new-finished", models.StatusFinished, recent, recent)

	admin := s.signedInAs("admin")
	_, body := admin.get("/admin/queue-watcher/jobs?status=running")
	s.Equal(2, strings.Count(body, `data-testid="run-row"`))

	response, _ := admin.post("/admin/queue-watcher/jobs/mark-stale", url.Values{})
	s.Equal("/admin/queue-watcher/jobs?saved=marked-stale", location(response))
	status := func(jobID string) string {
		var run models.QueueJob
		if facades.Orm().Query().Where("job_id", jobID).First(&run) != nil || run.ID == 0 {
			return "gone"
		}
		return run.Status
	}
	s.Equal(models.StatusStale, status("stuck"))
	s.Equal(models.StatusRunning, status("busy"), "a run of a few seconds is still running")

	response, _ = admin.post("/admin/queue-watcher/jobs/purge", url.Values{})
	s.Equal("/admin/queue-watcher/jobs?saved=purged", location(response))
	s.Equal("gone", status("old-finished"))
	s.Equal(models.StatusFinished, status("new-finished"))
	s.Equal(models.StatusRunning, status("busy"))

	s.Require().NoError(facades.Artisan().Call("queue-watcher:purge --days=1"))
}
