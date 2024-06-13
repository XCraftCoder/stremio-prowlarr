package pipe

const (
	defaultConcurrency = 5
)

type Pipe[R any] struct {
	source  Source[R]
	stages  []pipeStage[R]
	stopped chan struct{}
	errCh   chan error
}

type Source[R any] func() ([]R, error)
type Sink[R any] func(record R) error

type pipeStage[R any] struct {
	fn          func(r R) ([]R, error)
	concurrency int
}

type StageOption[R any] func(p *pipeStage[R])

func New[R any](source Source[R]) *Pipe[R] {
	return &Pipe[R]{
		source:  source,
		errCh:   make(chan error, 1),
		stopped: make(chan struct{}),
	}
}

func (p *Pipe[R]) Map(fn func(r R) (R, error), opts ...StageOption[R]) {
	p.FanOut(func(in R) ([]R, error) {
		out, err := fn(in)
		if err != nil {
			return nil, err
		}

		return []R{out}, nil
	}, opts...)
}

func (p *Pipe[R]) FanOut(fn func(r R) ([]R, error), opts ...StageOption[R]) {
	stage := pipeStage[R]{
		fn:          fn,
		concurrency: defaultConcurrency,
	}

	for _, opt := range opts {
		opt(&stage)
	}

	p.stages = append(p.stages, stage)
}

func (p *Pipe[R]) Sink(sink Sink[R]) error {
	outCh := p.startSource()
	for _, stage := range p.stages {
		outCh = p.startStage(stage, outCh)
	}
	p.startSink(sink, outCh)
	<-p.stopped
	return <-p.errCh
}

func (p *Pipe[R]) Stop() {
	select {
	case <-p.stopped:
	default:
		close(p.errCh)
		close(p.stopped)
	}
}

func (p *Pipe[R]) startSource() <-chan R {
	outCh := make(chan R)

	go func() {
		defer close(outCh)
		records, err := p.source()
		if err != nil {
			p.reportError(err)
			return
		}

		p.sendRecords(records, outCh)
	}()

	return outCh
}

func (p *Pipe[R]) startStage(stage pipeStage[R], inCh <-chan R) <-chan R {
	outCh := make(chan R)
	go func() {
		defer close(outCh)
		for r := range inCh {
			outs, err := stage.fn(r)
			if err != nil {
				p.reportError(err)
				return
			} else {
				p.sendRecords(outs, outCh)
			}
		}
	}()
	return outCh
}

func (p *Pipe[R]) startSink(sink Sink[R], inCh <-chan R) {
	go func() {
		for record := range inCh {
			err := sink(record)
			if err != nil {
				p.reportError(err)
			}
		}
		p.Stop()
	}()
}

func (p *Pipe[R]) reportError(err error) {
	select {
	case <-p.stopped:
	case p.errCh <- err:
		p.Stop()
	default:
	}
}

func (p *Pipe[R]) sendRecords(records []R, outCh chan<- R) {
	for _, record := range records {
		select {
		case <-p.stopped:
		case outCh <- record:
		}
	}
}
