// Copyright © 2022 99nil.
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

package sse

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/r3labs/sse/v2"
)

const ContentType = "text/event-stream"

var EOFMessage = NewMessage("", "error", "eof")

func NewMessage(id, event, data string) *Message {
	return &Message{ID: id, Event: event, Data: data}
}

func NewErrMessage(id, data string) *Message {
	return &Message{ID: id, Event: "error", Data: data}
}

type Message struct {
	ID    string
	Event string
	Data  string
}

func (m *Message) Send(w http.ResponseWriter) {
	if m.Data == "" {
		return
	}

	if m.ID != "" {
		fmt.Fprintf(w, "id: %s\n", m.ID)
	}
	if m.Event != "" {
		fmt.Fprintf(w, "event: %s\n", m.Event)
	}
	fmt.Fprintf(w, "data: %s\n\n", m.Data)
	w.(http.Flusher).Flush()
}

func Wrap(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Stream not supported", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", ContentType)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		w.WriteHeader(http.StatusOK)
		f.Flush()

		fn(w, r)
		EOFMessage.Send(w)
	}
}

func NewRequest(uri string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header = make(http.Header)
	req.Header.Set("Accept", ContentType)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	return req, nil
}

var NewEventStreamReader = sse.NewEventStreamReader

var (
	headerID    = []byte("id:")
	headerData  = []byte("data:")
	headerEvent = []byte("event:")
)

func TransferMessage(content []byte) (*Message, error) {
	if len(content) < 1 {
		return nil, errors.New("event message was empty")
	}

	var id, event, data []byte
	// Normalize the crlf to lf to make it easier to split the lines.
	content = bytes.Replace(content, []byte("\n\r"), []byte("\n"), -1)
	// Split the line by "\n" or "\r", per the spec.
	lineSet := bytes.FieldsFunc(content, func(r rune) bool { return r == '\n' || r == '\r' })
	for _, line := range lineSet {
		switch {
		case bytes.HasPrefix(line, headerID):
			id = append([]byte(nil), trimHeader(len(headerID), line)...)
		case bytes.HasPrefix(line, headerData):
			// The spec allows for multiple data fields per event, concatenated them with "\n".
			data = append(data[:], append(trimHeader(len(headerData), line), byte('\n'))...)
		// The spec says that a line that simply contains the string "data" should be treated as a data field with an empty body.
		case bytes.Equal(line, bytes.TrimSuffix(headerData, []byte(":"))):
			data = append(data, byte('\n'))
		case bytes.HasPrefix(line, headerEvent):
			event = append([]byte(nil), trimHeader(len(headerEvent), line)...)
		default:
			// Ignore any garbage that doesn't match what we're looking for.
		}
	}
	// Trim the last "\n" per the spec.
	data = bytes.TrimSuffix(data, []byte("\n"))
	return NewMessage(string(id), string(event), string(data)), nil
}

func trimHeader(size int, data []byte) []byte {
	if data == nil || len(data) < size {
		return data
	}

	data = data[size:]
	// Remove optional leading whitespace
	if len(data) > 0 && data[0] == 32 {
		data = data[1:]
	}
	// Remove trailing new line
	if len(data) > 0 && data[len(data)-1] == 10 {
		data = data[:len(data)-1]
	}
	return data
}
