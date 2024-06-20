package pipe

import "sync"

const (
	defaultChannelConcurrency = 10
)

type channelStage[R any] struct {
	fn          func(r *R, stopCh <-chan struct{}, outCh chan<- *R) error
	concurrency int
	reportError func(err error)
	stopped     <-chan struct{}
}

func (s *channelStage[R]) process(inCh <-chan *R, outCh chan<- *R) {
	defer close(outCh)

	wg := &sync.WaitGroup{}
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range inCh {
				err := s.fn(r, s.stopped, outCh)
				if err != nil {
					s.reportError(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func (s *channelStage[R]) getBufSize() int {
	return 0
}
