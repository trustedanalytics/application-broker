package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Env

func getServiceDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error getting dir: %v", err)
		return "./"
	}
	return pwd
}

func setEnv(key, val string) bool {
	err := os.Setenv(key, val)
	if err != nil {
		// don't print val as it may include secretes
		log.Printf("error setting env var: %s", key)
		return false
	}
	return true
}

func pathExists(path string) bool {
	if len(path) < 1 {
		log.Printf("null path: %s", path)
		return false
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Printf("path not found: %v", path)
		} else {
			log.Printf("error on path: %s -> %v", path, err)
		}
		return false
	}

	return true
}

// Password

func genRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	en := base64.URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	raw := string(d)
	return raw[0:length]
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

func getNowInUtc() time.Time {
	return time.Now().UTC()
}

func getTime(f string) string {
	if len(f) < 1 {
		f = time.RFC850
	}
	return fmt.Sprintln(getNowInUtc().Format(f))
}
