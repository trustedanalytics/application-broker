package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Env

func ParseInt(s string, d int) int {
	if len(s) < 1 {
		return d
	}
	//strconv.Btoi64
	v, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		log.Fatalf("unable to parse int from %s: %v", s, err)
		return d
	}
	return int(v)
}

func GetEnvVarAsString(k, d string) string {
	if len(k) < 1 {
		return d
	}
	s := os.Getenv(k)
	if len(s) < 1 {
		return d
	}
	return s
}

func GetEnvVarAsInt(k string, d int) int {
	s := GetEnvVarAsString(k, "")
	if len(s) < 1 {
		return d
	}
	v, err := strconv.ParseInt(s, d, 8)
	if err != nil {
		log.Fatalf("unable to parse int from %s: %v", k, err)
		return d
	}
	return int(v)
}

func GetEnvVarAsBool(k string, d bool) bool {
	s := GetEnvVarAsString(k, "")
	if len(s) < 1 {
		return d
	}
	v, err := strconv.ParseBool(k)
	if err != nil {
		log.Fatalf("unable to parse bool from %s: %v", k, err)
		return d
	}
	return v
}

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
