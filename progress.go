package progress

import (
	"os"
	"time"
)

// restore last progress after 200ms of silence
const defaultDelay = 200 * time.Millisecond

type _Job struct {
	isProgress bool
	data       string
	done       chan struct{}
}

type _Manager struct {
	jobCh                chan _Job
	timer                *time.Timer
	delay                time.Duration
	lastProgress         string
	lastOutputIsProgress bool // implies lastOutputInline == false, progress should not contain '\n'
	lastOutputInline     bool
	timerRunning         bool
}

func newManager() *_Manager {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	jobCh := make(chan _Job, 1)
	manager := &_Manager{
		jobCh: jobCh,
		timer: timer,
		delay: defaultDelay,
	}
	go manager.goroutine()
	return manager
}

func (m *_Manager) goroutine() {
	for {
		select {
		case job := <-m.jobCh:
			m.write(job.isProgress, job.data)
			job.done <- struct{}{}
		case <-m.timer.C:
			m.restoreProgress()
		}
	}
}

func (m *_Manager) write(isProgress bool, data string) {
	if isProgress {
		m.stopTimer()
		if data == "" {
			m.clearProgress()
			return
		}
	}
	if m.lastOutputIsProgress {
		os.Stderr.WriteString("\r\033[K")
	} else if isProgress && m.lastOutputInline {
		os.Stderr.WriteString("\n")
	}
	os.Stderr.WriteString(data)
	if isProgress {
		m.lastProgress = data
	} else {
		m.lastOutputInline = data[len(data)-1] != '\n'
		if m.lastProgress != "" {
			m.resetTimer()
		}
	}
	m.lastOutputIsProgress = isProgress
}

func (m *_Manager) clearProgress() {
	if m.lastOutputIsProgress {
		os.Stderr.WriteString("\r\033[K")
		m.lastOutputIsProgress = false
		m.lastOutputInline = false
	}
	m.lastProgress = ""
}

func (m *_Manager) restoreProgress() {
	if m.lastOutputInline {
		os.Stderr.WriteString("\n")
	}
	os.Stderr.WriteString(m.lastProgress)
	m.lastOutputIsProgress = true
	m.lastOutputInline = true
	m.timerRunning = false
}

func (m *_Manager) stopTimer() {
	if m.timerRunning {
		if !m.timer.Stop() {
			<-m.timer.C
		}
	}
	m.timerRunning = false
}

func (m *_Manager) resetTimer() {
	if m.timerRunning {
		if !m.timer.Stop() {
			<-m.timer.C
		}
	}
	m.timer.Reset(m.delay)
	m.timerRunning = true
}

func (m *_Manager) Write(isProgress bool, data string) {
	done := make(chan struct{}, 1)
	m.jobCh <- _Job{
		isProgress: isProgress,
		data:       data,
		done:       done,
	}
	<-done
}

var manager = newManager()

type _Writer struct{}

func (_Writer) WriteString(s string) (int, error) {
	manager.Write(false, s)
	return len(s), nil
}

func (_Writer) Write(s []byte) (int, error) {
	manager.Write(false, string(s))
	return len(s), nil
}

// You may want to wrap log output like this:
//  log.SetOutput(progress.Writer)
var Writer = _Writer{}

func Update(s string) {
	manager.Write(true, s)
}

func Clear() {
	manager.Write(true, "")
}
