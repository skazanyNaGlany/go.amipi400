package components

import (
	"time"

	"github.com/skazanyNaGlany/go.amipi400/interfaces"
)

type RunnersBlocker struct {
	runners []interfaces.Runner
}

func (rb *RunnersBlocker) AddRunner(runner interfaces.Runner) {
	rb.runners = append(rb.runners, runner)
}

func (rb *RunnersBlocker) BlockUntilRunning() {
	for {
		time.Sleep(time.Second * 1)

		runningCount := 0

		for _, runner := range rb.runners {
			if runner.IsRunning() {
				runningCount += 1
			}
		}

		if runningCount < len(rb.runners) {
			break
		}
	}
}
