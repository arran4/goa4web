package backgroundtaskworker

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/testhelpers"
)

type mockRoundTripper func(req *http.Request) *http.Response

func (f mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

type httpClientTestTask struct {
	tasks.TaskString
	TargetURL string
	executed  chan struct{}
}

var _ tasks.Task = (*httpClientTestTask)(nil)
var _ tasks.BackgroundTasker = (*httpClientTestTask)(nil)

func (t *httpClientTestTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	cd := common.FromContext(ctx)
	if cd == nil {
		return nil, nil
	}
	client := cd.HTTPClient()
	if client != nil {
		resp, err := client.Get(t.TargetURL)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}
	close(t.executed)
	return nil, nil
}

func TestWorker_PreservesHTTPClientThroughWorkerPath(t *testing.T) {
	bus := eventbus.NewBus()
	qs := testhelpers.NewQuerierStub()

	var clientCalled bool
	var calledURL string
	client := &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) *http.Response {
			clientCalled = true
			calledURL = req.URL.String()
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	workerFinished := make(chan struct{})
	go func() {
		Worker(ctx, bus, qs, nil, WithHTTPClient(client), WithReady(ready))
		close(workerFinished)
	}()
	<-ready

	task := &httpClientTestTask{
		TaskString: "test:http_client",
		TargetURL:  "https://example.com/test-endpoint",
		executed:   make(chan struct{}),
	}

	err := bus.Publish(eventbus.TaskEvent{
		Task:    task,
		Outcome: eventbus.TaskOutcomeSuccess,
		Time:    time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to publish task: %v", err)
	}

	select {
	case <-task.executed:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for task to execute")
	}

	if !clientCalled {
		t.Error("expected configured HTTP client to be called in Worker path")
	}
	if calledURL != "https://example.com/test-endpoint" {
		t.Errorf("expected called URL 'https://example.com/test-endpoint', got %q", calledURL)
	}

	cancel()
	select {
	case <-workerFinished:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit cleanly on cancel")
	}
}

type countingTask struct {
	tasks.TaskString
	id int
	wg *sync.WaitGroup
}

var _ tasks.Task = (*countingTask)(nil)
var _ tasks.BackgroundTasker = (*countingTask)(nil)

func (t *countingTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	t.wg.Done()
	return nil, nil
}

func TestWorker_BurstReliabilityWithoutJobLoss(t *testing.T) {
	bus := eventbus.NewBus()
	qs := testhelpers.NewQuerierStub()

	const burstCount = 250
	var wg sync.WaitGroup
	wg.Add(burstCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	workerFinished := make(chan struct{})
	go func() {
		Worker(ctx, bus, qs, nil, WithConcurrency(10), WithReady(ready))
		close(workerFinished)
	}()
	<-ready

	for i := 0; i < burstCount; i++ {
		task := &countingTask{
			TaskString: "test:count",
			id:         i,
			wg:         &wg,
		}
		err := bus.Publish(eventbus.TaskEvent{
			Task:    task,
			Outcome: eventbus.TaskOutcomeSuccess,
			Time:    time.Now(),
		})
		if err != nil {
			t.Fatalf("failed to publish task %d: %v", i, err)
		}
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
	}()

	select {
	case <-allDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for burst tasks to execute")
	}

	cancel()
	select {
	case <-workerFinished:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit cleanly")
	}
}

type concurrencyTask struct {
	tasks.TaskString
	active      *int32
	maxObserved *int32
	wg          *sync.WaitGroup
}

var _ tasks.Task = (*concurrencyTask)(nil)
var _ tasks.BackgroundTasker = (*concurrencyTask)(nil)

func (t *concurrencyTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	cur := atomic.AddInt32(t.active, 1)
	for {
		oldMax := atomic.LoadInt32(t.maxObserved)
		if cur <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt32(t.maxObserved, oldMax, cur) {
			break
		}
	}

	time.Sleep(10 * time.Millisecond)
	atomic.AddInt32(t.active, -1)
	t.wg.Done()
	return nil, nil
}

func TestWorker_BoundedConcurrency(t *testing.T) {
	bus := eventbus.NewBus()
	qs := testhelpers.NewQuerierStub()

	const concurrencyLimit = 3
	const taskCount = 30

	var active int32
	var maxObserved int32
	var wg sync.WaitGroup
	wg.Add(taskCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	workerFinished := make(chan struct{})
	go func() {
		Worker(ctx, bus, qs, nil, WithConcurrency(concurrencyLimit), WithReady(ready))
		close(workerFinished)
	}()
	<-ready

	for i := 0; i < taskCount; i++ {
		task := &concurrencyTask{
			TaskString:  "test:concurrency",
			active:      &active,
			maxObserved: &maxObserved,
			wg:          &wg,
		}
		err := bus.Publish(eventbus.TaskEvent{
			Task:    task,
			Outcome: eventbus.TaskOutcomeSuccess,
			Time:    time.Now(),
		})
		if err != nil {
			t.Fatalf("failed to publish task %d: %v", i, err)
		}
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
	}()

	select {
	case <-allDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for concurrency tasks to execute")
	}

	if observed := atomic.LoadInt32(&maxObserved); observed > concurrencyLimit {
		t.Errorf("expected max concurrency <= %d, observed %d", concurrencyLimit, observed)
	}

	cancel()
	select {
	case <-workerFinished:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit cleanly")
	}
}

type followUpInitialTask struct {
	tasks.TaskString
}

var _ tasks.Task = (*followUpInitialTask)(nil)
var _ tasks.BackgroundTasker = (*followUpInitialTask)(nil)

func (t *followUpInitialTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	return &followUpResultTask{TaskString: "test:follow_up_result"}, nil
}

type followUpResultTask struct {
	tasks.TaskString
}

var _ tasks.Task = (*followUpResultTask)(nil)

func TestWorker_FollowUpTaskPublishing(t *testing.T) {
	bus := eventbus.NewBus()
	qs := testhelpers.NewQuerierStub()

	subCh := bus.Subscribe(eventbus.TaskMessageType)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	go Worker(ctx, bus, qs, nil, WithReady(ready))
	<-ready

	err := bus.Publish(eventbus.TaskEvent{
		Task:    &followUpInitialTask{TaskString: "test:follow_up_initial"},
		Outcome: eventbus.TaskOutcomeSuccess,
		Time:    time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to publish initial task: %v", err)
	}

	// Drain initial and check for follow-up event
	var foundFollowUp bool
	timeout := time.After(2 * time.Second)
	for !foundFollowUp {
		select {
		case env := <-subCh:
			if evt, ok := env.Msg.(eventbus.TaskEvent); ok {
				if named, ok := evt.Task.(tasks.Name); ok && named.Name() == "test:follow_up_result" {
					foundFollowUp = true
				}
			}
			env.Ack()
		case <-timeout:
			t.Fatal("timed out waiting for follow-up task to be published")
		}
	}
}
