package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Password

func genPassword(key string, length int) string {
	var bytes = make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = key[b%byte(len(key))]
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// JSON

func toString(o interface{}) (string, error) {
	objStr, err := json.Marshal(o)
	if err != nil {
		log.Printf("unable to marshal: %v", o)
		log.Panicln(err)
		return "", err
	}
	return fmt.Sprintln(string(objStr)), nil
}

// Scheduling

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)
	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()
	return stop
}

// Command

type simpleCommand struct {
	command string
	args    string
	output  string
	err     error
}

func exeCmd(c *simpleCommand) {
	var wg sync.WaitGroup
	wg.Add(1)
	go exeCmdAsync(c, &wg)
	wg.Wait()
}

func exeCmdAsync(c *simpleCommand, wg *sync.WaitGroup) {
	// don't log the command, could include passwords
	if wg == nil {
		return
	}
	if c == nil {
		wg.Done()
	}
	cmd := exec.Command(c.command, c.args)
	out, err := cmd.CombinedOutput()
	c.err = err
	// yep, this is a hack to pass simple tests
	// expecting users to encode results
	c.output = strings.Trim(string(out), "\n")
	wg.Done()
}
