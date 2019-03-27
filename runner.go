/*
Copyright 2016 - Jaume Arús

Author Jaume Arús - jaumearus@gmail.com

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package runner

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Getting a runner instance
// Parameters :
//    tasks_checking_interval: Frequency in milliseconds of tasks checking done by the runner.
func GetRunner(tasks_checking_interval int64) (Runner_I, error) {

	var lcfg zap.Config = zap.NewProductionConfig()
	lcfg.EncoderConfig.EncodeTime = SyslogTimeEncoder
	logger, _ := lcfg.Build()

	return &runner{tasks_checking_interval: tasks_checking_interval, tasks: make(map[int64]*RunningTask), isRunnnig: false, logger: logger}, nil
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

type RunningTask struct {
	Task          Task_I
	Args          []interface{}
	FirstWakingUp time.Time
	RunningFrom   []time.Time
}

type Runner_I interface {
	Start() error
	Shutdown() error
	WakeUpTask(Task_I, ...interface{}) error
	IsTaskRunning(int64) (bool, error)
	GetRunningTasksIDs() ([]int64, error)
	FinishTaskById(int64) error
}

type runner struct {
	logger                  *zap.Logger
	tasks_checking_interval int64
	tasks                   map[int64]*RunningTask
	isRunnnig               bool
}

// Waking up a new task, that implements the Task_I interface
// Parameters:
// 		task: The task to be waked up
//		args: A list of one or more arguments which will be passed to the task implementation as parameters
func (r *runner) WakeUpTask(task Task_I, args ...interface{}) error {

	// Check if already exists a task with the same ID
	if _, ok := r.tasks[task.GetID()]; ok {
		return errors.New(fmt.Sprintf("Already exists a task with ID: %d", task.GetID()))
	}
	// Register the task
	task.SetLogger(r.logger)
	r.tasks[task.GetID()] = &RunningTask{Task: task, Args: args}
	return nil

}

func (r *runner) IsTaskRunning(id int64) (bool, error) {
	t, ok := r.tasks[id]
	if !ok {
		return false, errors.New(fmt.Sprintf("There is not any task with ID %d !!!", id))
	}
	return t.Task.IsRunning(), nil
}

func (r *runner) GetRunningTasksIDs() ([]int64, error) {
	ids := make([]int64, len(r.tasks))
	c := 0
	for id, _ := range r.tasks {
		ids[c] = id
		c += 1
	}
	return ids, nil
}

// Starting the runner. If this method is not called, the tasks wont be fire up
func (r *runner) Start() error {
	go func(r *runner) {
		t := time.NewTicker(time.Duration(r.tasks_checking_interval) * time.Millisecond)
		for _ = range t.C {
			if !r.isRunnnig {
				r.run()
			}
		}
	}(r)

	return nil
}

func (r *runner) Shutdown() error {
	var finisher TaskManager_I
	for _, task := range r.tasks {
		t := task
		finisher = &taskManager{t.Task}
		finisher.Finish()
	}
	return nil
}

func (r *runner) FinishTaskById(id int64) error {
	t, ok := r.tasks[id]
	if !ok {
		return errors.New(fmt.Sprintf("There is not any task with ID %d !!!", id))
	}
	finisher := &taskManager{t.Task}
	finisher.Finish()
	return nil
}

func (r *runner) run() {
	defer func() {
		r.isRunnnig = false
	}()

	t := time.NewTicker(time.Duration(r.tasks_checking_interval) * time.Millisecond)
	for _ = range t.C {
		r.checkTasks()
	}
}

// Checking the status of the tasks.
// This method is invoked with the frequency filled in the tasks_checking_interval of the method GetRunner()
func (r *runner) checkTasks() error {

	var t time.Time
	var runningTime int64
	var taskDuration int64

	for _, rtask := range r.tasks {
		t = time.Now()
		if rtask.RunningFrom == nil {
			// The task has not been woke up yet. Waking it up...
			r.startTaskRunning(rtask, t)

		} else {

			// Check task duration (task duration <= 0 stands for infinite task)
			runningTime = int64(t.Sub(rtask.FirstWakingUp).Seconds() * 1000)
			taskDuration = int64(rtask.Task.GetDuration().Seconds() * 1000)

			if rtask.Task.IsFinished() || (taskDuration > 0 && (runningTime >= taskDuration)) {
				r.ensureTaskFinish(rtask)
			} else {

				// Check if the task is running
				if !rtask.Task.IsRunning() {
					r.ensureTaskIsRunning(rtask, t)
				} else {
					// Keeping the task running ..
					//r.logger.Debug("keep the task running",zap.Int64("running_time",runningTime),zap.Int64("task_duration", taskDuration))
				}
			}
		}
	}

	return nil
}

func (r *runner) startTaskRunning(rtask *RunningTask, t time.Time) {
	rtask.FirstWakingUp = t
	rtask.RunningFrom = []time.Time{t}
	rtask.Task.Run(rtask.Args)
}

func (r *runner) ensureTaskFinish(rtask *RunningTask) {
	if rtask.Task.IsRunning() {
		// This finishes the task and fires up the response sending
		rtask.Task.Finalize()
	}
	delete(r.tasks, rtask.Task.GetID())
	return
}

func (r *runner) ensureTaskIsRunning(rtask *RunningTask, t time.Time) {
	rtask.RunningFrom = append(rtask.RunningFrom, t)
	rtask.Task.Run(rtask.Args)
}
