package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

import (
	"github.com/dubbogo/dubbo-go-proxy/pkg/common/constant"
	"github.com/dubbogo/dubbo-go-proxy/pkg/common/extension"
	"github.com/dubbogo/dubbo-go-proxy/pkg/config"
	"github.com/dubbogo/dubbo-go-proxy/pkg/context"
	ctxHttp "github.com/dubbogo/dubbo-go-proxy/pkg/context/http"
	"github.com/dubbogo/dubbo-go-proxy/pkg/model"
	"github.com/dubbogo/dubbo-go-proxy/pkg/router"
	"github.com/stretchr/testify/assert"
)

func getTestContext() *ctxHttp.HttpContext {
	l := ListenerService{
		Listener: &model.Listener{
			Name: "test",
			Address: model.Address{
				SocketAddress: model.SocketAddress{
					Protocol: model.HTTP,
					Address:  "0.0.0.0",
					Port:     8888,
				},
			},
			FilterChains: []model.FilterChain{},
		},
	}

	hc := &ctxHttp.HttpContext{
		Listener:              l.Listener,
		FilterChains:          l.FilterChains,
		HttpConnectionManager: l.findHttpManager(),
		BaseContext:           context.NewBaseContext(),
	}
	hc.ResetWritermen(httptest.NewRecorder())
	hc.Reset()
	return hc
}

func getMockAPI(verb config.HTTPVerb, urlPattern string) router.API {
	inbound := config.InboundRequest{}
	integration := config.IntegrationRequest{}
	method := config.Method{
		OnAir:              true,
		HTTPVerb:           verb,
		InboundRequest:     inbound,
		IntegrationRequest: integration,
	}
	return router.API{
		URLPattern: urlPattern,
		Method:     method,
	}
}
func TestRouteRequest(t *testing.T) {
	mockAPI := getMockAPI(config.MethodPost, "/mock/test")
	mockAPI.Method.OnAir = false

	apiDiscoverySrv := extension.GetMustApiDiscoveryService(constant.LocalMemoryApiDiscoveryService)
	apiDiscoverySrv.AddAPI(mockAPI)
	apiDiscoverySrv.AddAPI(getMockAPI(config.MethodGet, "/mock/test"))

	listener := NewDefaultHttpListener()
	listener.pool.New = func() interface{} {
		return getTestContext()
	}
	r := bytes.NewReader([]byte("test"))

	req, _ := http.NewRequest("GET", "/mock/test", r)
	ctx := listener.pool.Get().(*ctxHttp.HttpContext)
	api, err := listener.routeRequest(ctx, req)
	assert.Nil(t, err)
	assert.NotNil(t, api)
	assert.Equal(t, api.URLPattern, "/mock/test")
	assert.Equal(t, api.HTTPVerb, config.MethodGet)

	req, _ = http.NewRequest("GET", "/mock/test2", r)
	ctx = listener.pool.Get().(*ctxHttp.HttpContext)
	api, err = listener.routeRequest(ctx, req)
	assert.EqualError(t, err, "Requested URL /mock/test2 not found")
	assert.Equal(t, ctx.StatusCode(), 404)
	assert.Equal(t, api, router.API{})

	req, _ = http.NewRequest("POST", "/mock/test", r)
	ctx = listener.pool.Get().(*ctxHttp.HttpContext)
	api, err = listener.routeRequest(ctx, req)
	assert.EqualError(t, err, "Requested API POST /mock/test does not online")
	assert.Equal(t, ctx.StatusCode(), 406)
	assert.Equal(t, api, router.API{})
}
