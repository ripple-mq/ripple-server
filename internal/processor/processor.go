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

// Singleton class to get task executor
func GetProcessor() *Processor {
	if processorInstance != nil {
		return processorInstance
	}
	processorInstance = newProcessor()
	processorInstance.Run()
	return processorInstance
}

func newProcessor() *Processor {
	return &Processor{q: collection.NewConcurrentQueue[Task]()}
}

func (t *Processor) Run() {
	go t.run()
}

// Run continuously poll task from queue and execute asynchronously
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
