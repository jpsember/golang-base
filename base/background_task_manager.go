package base

import (
	"sync"
)

type taskInfoStruct struct {
	key               string
	lastExecutionTime int64
	period            int64
	task              func()
}

type taskInfo = *taskInfoStruct

const (
	bgtaskmgrState_new = iota
	bgtaskmgrState_started
	bgtaskmgrState_stopping
	bgtaskmgrState_stopped
)

type BackgroundTaskManagerStruct struct {
	state   int
	taskMap map[string]taskInfo
	lock    sync.RWMutex
}

var shared = NewBackgroundTaskManager()

func SharedBackgroundTaskManager() BackgroundTaskManager {
	return shared
}

type BackgroundTaskManager = *BackgroundTaskManagerStruct

func NewBackgroundTaskManager() BackgroundTaskManager {
	t := &BackgroundTaskManagerStruct{
		taskMap: make(map[string]taskInfo),
	}
	return t
}

func (b BackgroundTaskManager) Start() BackgroundTaskManager {
	CheckState(b.state == bgtaskmgrState_new)
	b.setState(bgtaskmgrState_started)
	go b.perform()
	return b
}

func (b BackgroundTaskManager) Stop() BackgroundTaskManager {
	if b.state == bgtaskmgrState_started {
		b.setState(bgtaskmgrState_stopping)
		b.executeTasksPriorToStopping()
	}
	b.setState(bgtaskmgrState_stopped)
	return b
}

func (b BackgroundTaskManager) executeTasksPriorToStopping() {
	b.lock.RLock()
	for _, v := range b.taskMap {
		currTime := CurrentTimeMs()
		v.lastExecutionTime = currTime
		b.executeTask(v)
	}
	b.lock.RUnlock()
}

func (b BackgroundTaskManager) Add(key string, period int, task func()) BackgroundTaskManager {
	b.lock.Lock()
	defer b.lock.Unlock()
	CheckState(!HasKey(b.taskMap, key))
	b.taskMap[key] = &taskInfoStruct{
		key:    key,
		period: int64(period),
		task:   task,
	}

	if b.state == bgtaskmgrState_new {
		b.Start()
	}
	return b
}

func (b BackgroundTaskManager) perform() {
	pr := PrIf("BackgroundTaskManager.perform", false)
	for b.state != bgtaskmgrState_stopped {
		pr("tick")
		SleepMs(200)
		b.lock.RLock()
		for _, v := range b.taskMap {
			currTime := CurrentTimeMs()
			elapsed := currTime - v.lastExecutionTime
			pr("currTime:", currTime, "lastExec:", v.lastExecutionTime, "elapsed:", elapsed)
			if elapsed >= v.period {
				v.lastExecutionTime = currTime
				pr("executing task:", v.key)
				b.executeTask(v)
			}
		}
		b.lock.RUnlock()
	}
}

func (b BackgroundTaskManager) setState(state int) {
	pr := PrIf("bgtaskmgr state", false)
	pr("state changing from", b.state, "to", state)
	b.state = state
}

func (b BackgroundTaskManager) executeTask(v taskInfo) {
	defer CatchPanic(func() {
		Pr("Caught panic executing task:", v.key)
	})
	v.task()
}
