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

package proxywasm

import (
	"mosn.io/proxy-wasm-go-host/common"

	proxywasm "mosn.io/proxy-wasm-go-host/proxywasm"
)

import (
	"github.com/apache/dubbo-go-pixiu/pkg/common/constant"
	"github.com/apache/dubbo-go-pixiu/pkg/common/extension/filter"
	"github.com/apache/dubbo-go-pixiu/pkg/context/http"
	"github.com/apache/dubbo-go-pixiu/pkg/wasm"
)

func init() {
	filter.RegisterHttpFilter(&Plugin{})
}

type importHandler struct {
	reqHeader common.HeaderMap
	proxywasm.DefaultImportsHandler
}

type (
	// Plugin manages wasm instances by pool.
	Plugin struct {
	}

	WasmFilterFactory struct {
		cfg *Config
	}

	WasmFilter struct {
		abiContextWrappers map[string]*wasm.ABIContextWrapper
	}

	Config struct {
		wasmServices []*WasmService `yaml:"wasm_services" json:"wasm_services" mapstructure:"wasm_services"`
	}

	WasmService struct {
		name string `yaml:"name" json:"name" mapstructure:"name"`
	}
)

func (w *WasmFilter) Decode(ctx *http.HttpContext) filter.FilterStatus {
	const HEADER = "header"

	// sample: display http header
	wrapper := w.abiContextWrappers[HEADER]
	wrapper.Context.Instance.Lock(ctx)
	defer wrapper.Context.Instance.Lock(ctx)

	wrapper.Context.GetExports().ProxyOnContextCreate(wrapper.ContextID, wasm.GetServiceRootID(HEADER))

	_, _ = wrapper.Context.GetExports().ProxyOnRequestHeaders(wrapper.ContextID, int32(len(ctx.Request.Header)), 1)

	_ = wrapper.Context.GetExports().ProxyOnDelete(wrapper.ContextID)

	return filter.Continue
}

func (w *WasmFilter) Encode(ctx *http.HttpContext) filter.FilterStatus {
	for _, wrapper := range w.abiContextWrappers {
		wasm.ContextDone(wrapper)
	}
	return filter.Continue
}

func (factory *WasmFilterFactory) Config() interface{} {
	return factory.cfg
}

func (factory *WasmFilterFactory) Apply() error {
	return nil
}

func (factory *WasmFilterFactory) PrepareFilterChain(ctx *http.HttpContext, chain filter.FilterChain) error {
	filter := &WasmFilter{
		abiContextWrappers: make(map[string]*wasm.ABIContextWrapper),
	}
	for _, service := range factory.cfg.wasmServices {
		filter.abiContextWrappers[service.name] = wasm.CreateABIContextByName(service.name, ctx)
	}
	chain.AppendEncodeFilters(filter)
	return nil
}

func (p *Plugin) Kind() string {
	return constant.HTTPWasmFilter
}

func (p *Plugin) CreateFilterFactory() (filter.HttpFilterFactory, error) {
	return &WasmFilterFactory{
		cfg: &Config{},
	}, nil
}
