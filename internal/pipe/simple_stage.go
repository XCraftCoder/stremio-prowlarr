package pipe

import "sync"

type simpleStage[R any] struct {
	fn          func(r *R) ([]*R, error)
	concurrency int
	reportError func(err error)
	stopped     <-chan struct{}
}

type SimpleStageOption[R any] func(p *simpleStage[R])

func (s *simpleStage[R]) process(inCh <-chan *R, outCh chan<- *R) {
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
					sendRecords(outs, outCh, s.stopped)
				}
			}
		}()
	}
	wg.Wait()
}
