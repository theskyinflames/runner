# Task runner

see [runner_test.go](./runner_test.go) to use cases

## Usage
1. Take the runner and start it
```go
    // Take the runner
    checkIntervalMS := int64(100)
	myrunner, err := GetRunner(checkIntervalMS, logger)
	if err != nil {
		return err
	}
	// Starting the runner
	err = myrunner.Start()
	if err != nil {
		t.Error(err)
	}
	
	t.Logf("Runner started \n")
```

2. Define the task
```go
mytaskimplementation = func(eot chan struct{}, taskManager TaskManager_I, args []interface{}) ([]interface{}, error) {
		
		var c int = 0
		var sleep_time int64 = args[0].(int64)
		
		var t time.Time
		var exitValue = 1000
		

		for { // This loop is necessary to be able to catch the ending signal from the eot (EndOfTask) channel
			select {
			case <-eot: // If a signal is read from this channel, the task will finish
				fmt.Println("The task finalizes by EOT signal ...")
				return []interface{}{c, t}, nil
				
			default: // The task implementation
				time.Sleep(time.Duration(sleep_time) * time.Millisecond)
				c += 1
				t = time.Now()
				
				if c == exitValue{
				    break
				}
			}
		}
		
		fmt.Println("The task has finished by its self ....")
		return []interface{}{c, t}, nil
	}
```

3. Get the runner task. It this case is an infinite task, in other words, a task without deadline
```go
    taskID := getUUID()
    mytask, err = GetTask(taskID, INFINITE_TASK_DURATION, mytaskimplementation, logger)
```

4. Waking up the task with a parameter and wating for the return
```go
    parameter := int64(100)
    err := myrunner.WakeUpTask(mytask, parameter)
```

5. Waiting for the response
```go
    response := <-mytask.GetResponseChan()
```
