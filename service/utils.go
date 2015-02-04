package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParseInt tries to parse a string into an int; else returns a default value
func ParseInt(s string, defaultInt int) int {
	if len(s) < 1 {
		return defaultInt
	}
	//strconv.Btoi64
	v, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		log.Fatalf("unable to parse int from %s: %v", s, err)
		return defaultInt
	}
	return int(v)
}

// GetEnvVarAsString gets an environment variable, or returns a default value if missing/empty
func GetEnvVarAsString(k, defaultEnvValue string) string {
	if len(k) < 1 {
		return defaultEnvValue
	}
	s := os.Getenv(k)
	if len(s) < 1 {
		return defaultEnvValue
	}
	return s
}

// GetEnvVarAsInt gets an env variable and parses to an int; or returns
// a default int if variable missing or not an int
func GetEnvVarAsInt(k string, defaultInt int) int {
	s := GetEnvVarAsString(k, "")
	if len(s) < 1 {
		return defaultInt
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("unable to parse int from %s: %v", k, err)
		return defaultInt
	}
	return int(v)
}

// GetEnvVarAsBool gets an env variable and parses to a bool; or returns
// a default bool if variable missing or not a bool
func GetEnvVarAsBool(k string, defaultBool bool) bool {
	s := GetEnvVarAsString(k, "")
	if len(s) < 1 {
		return defaultBool
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatalf("unable to parse bool from %s: %v", k, err)
		return defaultBool
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

	currDir, err := os.Getwd()
	if err != nil {
		log.Printf("%v", err)
		return false
	}
	log.Printf("current dir: %s", currDir)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Printf("path not found: %s", path)
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

func reduceInstanceID(id string) string {
	idSplit := strings.Split(id, "-")
	return strings.Join(idSplit[0:len(idSplit)-1], "-")
}
