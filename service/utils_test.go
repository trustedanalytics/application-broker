package service

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomPassword(t *testing.T) {

	pas := genRandomString(4)

	assert.Nil(t, pas, "nil password")
	assert.Equal(t, len(pas), 4, "wrong password length")
}

func TestSimpleCommand(t *testing.T) {

	cmd := &simpleCommand{
		command: "echo",
		args:    []string{"test"},
	}

	exeCmd(cmd)

	assert.Nil(t, cmd.err, "command failed")
	assert.NotNil(t, cmd.output, "nil output")
	assert.Equal(t, cmd.output, "test", "wrong output")
}

type MarshalingTestObj struct {
	S string
	N int
	B bool
}

func TestMarshaling(t *testing.T) {

	o := &MarshalingTestObj{S: "test", N: 1, B: true}
	s, err := toString(o)

	assert.Nil(t, err, "error on marshaling")
	assert.NotEmpty(t, s, "empty marshaling output")

	o2 := &MarshalingTestObj{}
	err2 := json.Unmarshal([]byte(s), &o2)

	assert.Nil(t, err2, "error on unmarshaling")
	assert.NotNil(t, o2, "nil unmarshaled from")
	assert.Equal(t, o.S, o2.S, "s not the same from: "+s)
	assert.Equal(t, o.N, o2.N, "n not the same from: "+s)
	assert.Equal(t, o.B, o2.B, "b not the same from: "+s)

}
