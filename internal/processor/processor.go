package processor

import (
	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/pkg/utils/collection"
)

type Task interface {
	Exec() error
}

type Processor struct {
	q *collection.ConcurrentQueue[Task]
}

var processorInstance *Processor

func GetProcessor() *Processor {
	if processorInstance != nil {
		return processorInstance
	}
	processorInstance = newProcessor()
	return processorInstance
}

func newProcessor() *Processor {
	return &Processor{q: collection.NewConcurrentQueue[Task]()}
}

func (t *Processor) Run() {
	go t.run()
}

func (t *Processor) run() {
	for {
		if t.q.IsEmpty() {
			continue
		}
		task := t.q.Poll()
		if err := task.Exec(); err != nil {
			log.Errorf("error while processing task: %v", err)
		}
	}
}

func (t *Processor) Add(task Task) {
	t.q.Push(task)
}
