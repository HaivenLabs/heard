package app

import (
	"context"
	"errors"
	"testing"
)

func TestWorkerDrainsAvailableOutboxEventsInOneBoundedBatch(t *testing.T) {
	processor := &stubOutboxProcessor{results: []outboxProcessResult{{processed: true}, {processed: true}, {processed: true}, {processed: false}}}
	worker := NewWorker(Config{WorkerBatchSize: 50}, processor)

	result := worker.processBatch(context.Background())
	if result.Attempted != 3 || result.Succeeded != 3 || result.Failed != 0 {
		t.Fatalf("unexpected batch result: %#v", result)
	}
}

func TestWorkerBatchNeverExceedsConfiguredBound(t *testing.T) {
	processor := &stubOutboxProcessor{alwaysProcessed: true}
	worker := NewWorker(Config{WorkerBatchSize: 2}, processor)

	result := worker.processBatch(context.Background())
	if result.Attempted != 2 || processor.calls != 2 {
		t.Fatalf("batch was not bounded: result=%#v calls=%d", result, processor.calls)
	}
}

func TestWorkerBatchCountsCommittedProcessingFailures(t *testing.T) {
	processor := &stubOutboxProcessor{results: []outboxProcessResult{{processed: true, err: errors.New("provider unavailable")}, {processed: false}}}
	worker := NewWorker(Config{WorkerBatchSize: 10}, processor)

	result := worker.processBatch(context.Background())
	if result.Attempted != 1 || result.Failed != 1 || result.Succeeded != 0 {
		t.Fatalf("unexpected failure counts: %#v", result)
	}
}

type outboxProcessResult struct {
	processed bool
	err       error
}

type stubOutboxProcessor struct {
	results         []outboxProcessResult
	alwaysProcessed bool
	calls           int
}

func (s *stubOutboxProcessor) ProcessNextOutboxEvent(context.Context) (bool, error) {
	s.calls++
	if s.alwaysProcessed {
		return true, nil
	}
	if len(s.results) == 0 {
		return false, nil
	}
	result := s.results[0]
	s.results = s.results[1:]
	return result.processed, result.err
}
