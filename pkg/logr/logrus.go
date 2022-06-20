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

package logr

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func NewLogrusInstance(cfg *Config) *logrusInstance {
	l := logrus.New()
	if cfg != nil {
		lvl, err := logrus.ParseLevel(cfg.Level)
		if err == nil {
			l.SetLevel(lvl)
		}
		if strings.TrimSpace(cfg.Filename) != "" {
			l.SetOutput(cfg)
		}
		l.SetFormatter(&logrus.TextFormatter{
			ForceColors:            true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			FullTimestamp:          true,
			TimestampFormat:        "2006/01/02 15:04:05",
		})
	}
	entry := logrus.NewEntry(l)
	return &logrusInstance{entry}
}

type logrusInstance struct {
	*logrus.Entry
}

func (l *logrusInstance) Level() LevelType {
	return l.Entry.Logger.Level.String()
}

func (l *logrusInstance) WithField(key string, val interface{}) Interface {
	entry := l.Entry.WithField(key, val)
	return &logrusInstance{entry}
}

func (l *logrusInstance) WithFields(fields map[string]interface{}) Interface {
	entry := l.Entry.WithFields(fields)
	return &logrusInstance{entry}
}

func (l *logrusInstance) WithError(err error) Interface {
	entry := l.Entry.WithField("error", err)
	return &logrusInstance{entry}
}
