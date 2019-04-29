package concurrent

import (
	"context"
	"sync"
)

type WorkerFunction func(context.Context, interface{}) error

const (
	OverFlowDrop = iota
	OverFlowReplace
	OverFlowNolimit
)

type ModuleOptions struct {
	FetchChanCount    int
	FetchBufferLen    int
	HandleError       bool
	OverflowBehaivour int //TODO
	WorkerCount       int
	F                 WorkerFunction
}

type BufferedModule struct {
	dataChan   chan interface{}
	dataBuffer []interface{}
	ctx        context.Context
	cancel     context.CancelFunc
	errChan    chan error
	opts       *ModuleOptions

	mux sync.Mutex
}

func (m *BufferedModule) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	m.mux.Lock()
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mux.Unlock()

	pubChan := make(chan interface{})
	defer close(pubChan)

	var wg sync.WaitGroup
	defer wg.Wait()

	for i := 0; i < m.opts.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-m.ctx.Done():
					return
				case d := <-pubChan:
					err := m.opts.F(ctx, d)
					if m.errChan != nil {
						m.errChan <- err
					}
				}
			}
		}()
	}

	var first interface{}
	var _pubChan chan interface{}
	for {
		if len(m.dataBuffer) > 0 {
			first = m.dataBuffer[0]
			_pubChan = pubChan
		} else {
			_pubChan = nil
		}

		select {
		case <-m.ctx.Done():
			return
		case data := <-m.dataChan:
			if len(m.dataBuffer) > m.opts.FetchBufferLen {
				switch m.opts.OverflowBehaivour {
				case OverFlowDrop:
					continue
				case OverFlowReplace:
					m.dataBuffer = m.dataBuffer[1:]
				case OverFlowNolimit:
					//do nothing, just append the data buffer
				}
			}
			m.dataBuffer = append(m.dataBuffer, data)
		case _pubChan <- first:
			m.dataBuffer = m.dataBuffer[1:]
		}
	}
}

func (m *BufferedModule) Stop() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.cancel != nil {
		m.cancel()
	}

	return nil
}

func (m *BufferedModule) ErrChan() <-chan error {
	return m.errChan
}

func (m *BufferedModule) Feed(data interface{}) {
	m.dataChan <- data
}

func NewBufferModule(opts ModuleOptions) (*BufferedModule, error) {
	m := &BufferedModule{
		dataChan:   make(chan interface{}, opts.FetchChanCount),
		dataBuffer: make([]interface{}, 0, opts.FetchBufferLen),
		opts:       &opts,
	}
	if opts.HandleError {
		m.errChan = make(chan error, m.opts.WorkerCount)
	}

	return m, nil
}
