package worker

import "log"

type worker struct {
}

func NewWorkers() *worker {
	return &worker{}
}

func (w *worker) Run() {
	log.Println("not implement")
}
