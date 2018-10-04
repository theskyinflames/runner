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
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

var err error

var myrunner Runner_I
var mytask Task_I
var mytaskimplementation TASK_IMPLEMENTATION

var logger *zap.Logger = zap.NewExample()

func TestRunner_Start(t *testing.T) {

	// Take the runner
	if myrunner, err = GetRunner(100, logger); err != nil {
		t.Error(err)
	} else {
		// Starting the runner
		if err = myrunner.Start(); err != nil {
			t.Error(err)
		} else {
			t.Logf("Runner started \n")
		}
	}
}

// Test for a finite task which is executed during all of its duration
func TestRunner_WakeUpTaskDuration(t *testing.T) {

	// Define the task implementation
	mytaskimplementation = func(eot chan struct{}, taskManager TaskManager_I, args []interface{}) ([]interface{}, error) {
		var c int = 0
		var sleep_time int64 = args[0].(int64)
		var t time.Time

		for {
			select {
			case <-eot:
				fmt.Println("The task finalizes by EOT signal ...")
				return []interface{}{c, t}, nil
			default:
				time.Sleep(time.Duration(sleep_time) * time.Millisecond)
				c += 1
				t = time.Now()
			}
		}
		fmt.Println("The task has finished by its self ....")
		return []interface{}{c, t}, nil
	}

	// Get the task
	if mytask, err = GetTask(1, time.Duration(3)*time.Second, mytaskimplementation, logger); err != nil {
		t.Error(err)
	} else {
		t.Logf("The task has been woken up\n")

		// Wakeup the task
		if err = myrunner.WakeUpTask(mytask, int64(100)); err != nil {
			t.Error(err)
		} else {

			// Take the response
			response := <-mytask.GetResponseChan()

			// Check for the response
			if response.Err != nil {
				t.Error(response.Err)
			} else {
				t.Logf("The result is : %#v\n", response.Result)
			}
		}
	}
}

// Test for a finite task which finishes before its maximum duration
func TestRunner_WakeUpTaskFinite(t *testing.T) {

	// Define the task implementation
	mytaskimplementation = func(eot chan struct{}, taskManager TaskManager_I, args []interface{}) ([]interface{}, error) {
		var c int = 0
		var sleep_time int64 = args[0].(int64)
		var t time.Time
		var goon bool = true

		for goon {
			select {
			case <-eot:
				fmt.Println("The task finalizes by EOT signal ...")
				return []interface{}{c, t}, nil
			default:
				time.Sleep(time.Duration(sleep_time) * time.Millisecond)
				c += 1
				t = time.Now()
				if c == 3 {
					// Finishing the task
					goon = false
				}
			}
		}
		fmt.Println("The task has finished by its self ....")

		// Says the runner this task must be finished
		taskManager.Finish()

		return []interface{}{c, t}, nil
	}

	// Get the task
	if mytask, err = GetTask(2, time.Duration(10)*time.Second, mytaskimplementation, logger); err != nil {
		t.Error(err)
	} else {
		t.Logf("The task has been woken up\n")

		// Wakeup the task
		if err = myrunner.WakeUpTask(mytask, int64(100), logger); err != nil {
			t.Error(err)
		} else {

			// Take the response
			response := <-mytask.GetResponseChan()

			// Check for the response
			if response.Err != nil {
				t.Error(response.Err)
			} else {
				t.Logf("The result is : %#v\n", response.Result)
			}
		}
	}
}

// Test for a finite task that breaks down and is resumed several times during its executin time
func TestRunner_WakeUpTaskResuming(t *testing.T) {

	// Define the task implementation
	mytaskimplementation = func(eot chan struct{}, taskManager TaskManager_I, args []interface{}) ([]interface{}, error) {
		var c int = 0
		var sleep_time int64 = args[0].(int64)
		var t time.Time
		var goon bool = true

		fmt.Printf("Resuming task started ...\n")

		for goon {
			select {
			case <-eot:
				fmt.Println("The task finalizes by EOT signal ...")
				return []interface{}{c, t}, nil
			default:
				time.Sleep(time.Duration(sleep_time) * time.Millisecond)
				c += 1
				t = time.Now()
				if c == 3 {
					//forces the flush of the task
					taskManager.Flush()
					panic("ooppp 0!!!")
				}
			}
		}
		fmt.Println("The task has finished by its self ....")
		return []interface{}{c, t}, nil
	}

	// Get the task
	if mytask, err = GetTask(3, time.Duration(3)*time.Second, mytaskimplementation, logger); err != nil {
		t.Error(err)
	} else {
		t.Logf("The task has been woken up\n")

		// Wakeup the task
		if err = myrunner.WakeUpTask(mytask, int64(100)); err != nil {
			t.Error(err)
		} else {

			_now := time.Now()
			results := make([]*TaskResponse, 0)
			for {
				results = append(results, <-mytask.GetResponseChan())
				if time.Now().Sub(_now).Seconds() > 3 {
					break
				}
			}

			// Check for the response
			t.Logf("The result is : %#v\n", results)
		}
	}
}
