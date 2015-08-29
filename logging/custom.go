/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	log "github.com/cihub/seelog"
	"os"
)

func Initialize() {
	log.RegisterReceiver("stderr", &StdErrReceiver{})

	logger, err := log.LoggerFromConfigAsFile("logger.config")
	if err != nil {
		println("Logger configuration failed! Err:" + err.Error())
	}
	log.ReplaceLogger(logger)
}

type StdErrReceiver struct {
}

func (ar *StdErrReceiver) ReceiveMessage(message string, level log.LogLevel, context log.LogContextInterface) error {
	os.Stderr.WriteString(message)
	return nil
}

func (ar *StdErrReceiver) AfterParse(initArgs log.CustomReceiverInitArgs) error { return nil }
func (ar *StdErrReceiver) Flush()                                               {}
func (ar *StdErrReceiver) Close() error                                         { return nil }
