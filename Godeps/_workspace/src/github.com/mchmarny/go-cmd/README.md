# go-cmd

Your basic exe command implementation

## Features 

* Command chaining
* Persistent Environment variables
* Sync/Async support  

## Installation 

To install go-cmd, use go get:

    go get github.com/mchmarny/go-cmd
    
To update go-cmd, use go get -u:

    go get -u github.com/mchmarny/go-cmd  

This will then make the following package available:

    github.com/mchmarny/go-cmd/cmd

## Example 

```
package main

import (
  "github.com/mchmarny/go-cmd"
  "log"
)

func main() {

    c := cmd.New("echo").
	         WithEnv("MY_VAR", "something").
	         WithArgs("-n").
	         WithArgs(testVal).
	         Exec()
	         
	if cmd.Err != nil {
		log.Fatalf("command failed: %v", cmd)
	}

	log.Printf("command output: %s", cmd.Out)

}
```