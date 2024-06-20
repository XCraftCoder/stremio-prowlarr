package pipe

import "sync"

type simpleStage[R any] struct {
	fn          func(r *R) ([]*R, error)
	concurrency int
	reportError func(err error)
	stopped     <-chan struct{}
}

type SimpleStageOption[R any] func(p *simpleStage[R])

func Concurrency[R any](concurrency int) SimpleStageOption[R] {
	return func(p *simpleStage[R]) {
		p.concurrency = concurrency
	}
}

func (s *simpleStage[R]) process(inCh <-chan *R, outCh chan<- *R) {
	defer close(outCh)

	wg := &sync.WaitGroup{}
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range inCh {
				outs, err := s.fn(r)
				if err != nil {
					s.reportError(err)
					return
				} else {
					SendRecords(outs, outCh, s.stopped)
				}
			}
		}()
	}
	wg.Wait()
}

func (s *simpleStage[R]) getBufSize() int {
	return 0
}
