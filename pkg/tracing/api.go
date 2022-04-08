/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tracing

import (
	"context"
)

import (
	"go.opentelemetry.io/otel/trace"
)

// Trace interface need to be implemented to construct your Tracer.
type Trace interface {
	// GetId gets ID string
	GetID() string
	// StartSpan creates new root span.
	StartSpan(name string, request interface{}) (context.Context, trace.Span)
	// StartSpanFromContext creates subSpan.
	StartSpanFromContext(name string, tx context.Context) (context.Context, trace.Span)
	// Close deletes trace from holder.
	Close()
}
