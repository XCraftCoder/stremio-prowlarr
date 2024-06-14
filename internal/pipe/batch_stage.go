package pipe

import "sync"

const (
	defaultBatchSize = 10
)

type batchStage[R any] struct {
	fn          func(r []R) ([]R, error)
	concurrency int
	batchSize   int
	reportError func(err error)
	stoped      <-chan struct{}
	batchCh     chan []R
}

type BatchStageOption[R any] func(p *batchStage[R])

func WorkerSize[R any](workerSize int) BatchStageOption[R] {
	return func(p *batchStage[R]) {
		p.concurrency = workerSize
	}
}

func (s *batchStage[R]) process(inCh <-chan R, sendRecords func(records []R)) {
	wg := &sync.WaitGroup{}
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range s.batchCh {
				outs, err := s.fn(batch)
				if err != nil {
					s.reportError(err)
					return
				} else {
					sendRecords(outs)
				}
			}
		}()
	}

	wg.Add(1)
	go s.batchRecords(wg, inCh)

	wg.Wait()
}

func (s *batchStage[R]) batchRecords(wg *sync.WaitGroup, inCh <-chan R) {
	defer wg.Done()
	for {
		select {
		case <-s.stoped:
			return
		default:
			select {
			case record := <-inCh:
				s.processNextBatch(record, inCh)
			case <-s.stoped:
				return
			}
		}
	}
}

func (s *batchStage[R]) processNextBatch(r R, inCh <-chan R) {
	newBatch := make([]R, 0, s.batchSize)
	newBatch = append(newBatch, r)
	for {
		if len(newBatch) == s.batchSize {
			select {
			case <-s.stoped:
			default:
				// not stopped, try to queue the next batch
				select {
				case s.batchCh <- newBatch:
				case <-s.stoped:
				}
			}
			return
		}

		select {
		case record := <-inCh:
			newBatch = append(newBatch, record)
		default:
			select {
			case record := <-inCh:
				newBatch = append(newBatch, record)
			case s.batchCh <- newBatch:
				return
			case <-s.stoped:
				return
			}
		}
	}
}
