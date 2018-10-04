# Task runner

see [runner_test.go](./runner_test.go) to use cases

## Usage
1. Take the runner
```go
    // Take the runner
    checkIntervalMS = int64(100)
	if myrunner, err = GetRunner(checkIntervalMS, logger); err != nil {
		t.Error(err)
	} else {
		// Starting the runner
		if err = myrunner.Start(); err != nil {
			t.Error(err)
		} else {
			t.Logf("Runner started \n")
		}
	}
```

2. Define the task
```go
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
```

3. Get the running task
```go
    mytask, err = GetTask(1, INFINITE_TASK_DURATION, mytaskimplementation, logger)
```

4. Wake the task with a parameter
```go
    parameter := int64(100)
    err := myrunner.WakeUpTask(mytask, parameter)
```

5. Waiting for the response
```go
    response := <-mytask.GetResponseChan()
```