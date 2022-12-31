// Copyright 2022 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grayzero

import (
	"fmt"
	"io"

	"github.com/aphistic/golf"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog"
)

type GraylogWriter struct {
	logger      *golf.Logger
	client      *golf.Client
	nextWriter  io.Writer
	serviceName string
	instanceID  string
}

// NewGraylogWriter returns new writer
func NewGraylogWriter(url string, nextWriter io.Writer, serviceName, instanceID string) (*GraylogWriter, error) {
	c, err := golf.NewClient()
	if err != nil {
		return nil, err
	}

	err = c.Dial(url)
	if err != nil {
		return nil, err
	}

	l, err := c.NewLogger()
	if err != nil {
		return nil, err
	}

	return &GraylogWriter{
		logger:      l,
		client:      c,
		nextWriter:  nextWriter,
		serviceName: serviceName,
		instanceID:  instanceID,
	}, nil
}

// Write extracts the message from the JSON and sends it in GELF format to GrayLog
func (w *GraylogWriter) Write(p []byte) (n int, err error) {
	var evt map[string]any
	err = jsonw.Unmarshal(p, &evt)
	if err != nil {
		return n, fmt.Errorf("cannot decode event: %w", err)
	}

	var fields = make(map[string]any, len(evt)+2)
	if w.serviceName != "" {
		fields["_service"] = w.serviceName
	}
	if w.instanceID != "" {
		fields["_instance"] = w.instanceID
	}
	var message string
	var level string
	for k, v := range evt {
		switch k {
		case zerolog.LevelFieldName:
			level = v.(string)
		case zerolog.MessageFieldName:
			message = v.(string)
		case zerolog.TimestampFieldName:
			// message timestamp will be recreated by client.QueueMsg(...)
			continue
			//case zerolog.CallerFieldName:
			//	// do nothing
		}
		fields[k] = v
	}

	newMsg := w.logger.NewMessage()
	newMsg.ShortMessage = message
	newMsg.Attrs = fields

	switch level {
	case "trace":
		newMsg.Level = golf.LEVEL_DBG
	case "debug":
		newMsg.Level = golf.LEVEL_DBG
	case "info":
		newMsg.Level = golf.LEVEL_INFO
	case "warn":
		newMsg.Level = golf.LEVEL_WARN
	case "error":
		newMsg.Level = golf.LEVEL_ERR
	case "fatal":
		newMsg.Level = golf.LEVEL_CRIT
	case "panic":
		newMsg.Level = golf.LEVEL_EMERG
	default:
		newMsg.Level = golf.LEVEL_NOTICE
	}

	if err := w.client.QueueMsg(newMsg); err != nil {
		return 0, err
	}

	if w.nextWriter != nil {
		return w.nextWriter.Write(p)
	} else {
		return len(p), nil
	}
}

// Close closes GrayLog client
func (w *GraylogWriter) Close() error {
	return w.client.Close()
}
