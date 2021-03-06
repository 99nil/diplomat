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
	"github.com/99nil/gopkg/logger"

	"github.com/natefinch/lumberjack"
)

type Config struct {
	//nolint
	lumberjack.Logger `json:",inline,squash"`

	Level string `json:"level"` // info, warn, error, debug
}

type Interface interface {
	logger.UniversalInterface

	Level() LevelType
	WithField(key string, val interface{}) Interface
	WithFields(fields map[string]interface{}) Interface
	WithError(err error) Interface
}

var std Interface = NewLogrusInstance(nil)

func Level() LevelType {
	return std.Level()
}

func SetDefault(l Interface) {
	std = l
}

func StdLogger() Interface {
	return std
}

func WithError(err error) Interface {
	return std.WithError(err)
}

func WithField(key string, val interface{}) Interface {
	return std.WithField(key, val)
}

func WithFields(fields map[string]interface{}) Interface {
	return std.WithFields(fields)
}

func Info(v ...interface{}) {
	std.Info(v...)
}

func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

func Warn(v ...interface{}) {
	std.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	std.Warnf(format, v...)
}

func Error(v ...interface{}) {
	std.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

func Debug(v ...interface{}) {
	std.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}
