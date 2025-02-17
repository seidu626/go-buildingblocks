// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package logging

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

type ZapAdapter struct {
	zl *zap.Logger
}

func NewZapAdapter(zapLogger *zap.Logger) *ZapAdapter {
	return &ZapAdapter{
		// Skip one call frame to exclude zap_adapter itself.
		// Or it can be configured when logger is created (not always possible).
		zl: zapLogger.WithOptions(zap.AddCallerSkip(1)),
	}
}

func (log *ZapAdapter) fields(keyValues []interface{}) []zap.Field {
	if len(keyValues)%2 != 0 {
		return []zap.Field{zap.Error(fmt.Errorf("odd number of keyValues pairs: %v", keyValues))}
	}

	var fields []zap.Field
	for i := 0; i < len(keyValues); i += 2 {
		key, ok := keyValues[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyValues[i])
		}
		fields = append(fields, zap.Any(key, keyValues[i+1]))
	}

	return fields
}

func (log *ZapAdapter) Debug(msg string, keyValues ...interface{}) {
	log.zl.Debug(msg, log.fields(keyValues)...)
}

func (log *ZapAdapter) Info(msg string, keyValues ...interface{}) {
	log.zl.Info(msg, log.fields(keyValues)...)
}

func (log *ZapAdapter) Warn(msg string, keyValues ...interface{}) {
	log.zl.Warn(msg, log.fields(keyValues)...)
}

func (log *ZapAdapter) Error(msg string, keyValues ...interface{}) {
	log.zl.Error(msg, log.fields(keyValues)...)
}

func (log *ZapAdapter) With(keyValues ...interface{}) log.Logger {
	return &ZapAdapter{zl: log.zl.With(log.fields(keyValues)...)}
}
