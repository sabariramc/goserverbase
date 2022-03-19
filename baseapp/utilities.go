package baseapp

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"

	"sabariram.com/goserverbase/constant"
	"sabariram.com/goserverbase/log"

	"github.com/google/uuid"
)

func (b *BaseApp) GetHttpCorrelationParams(r *http.Request) *log.CorrelationParmas {
	correlationId := r.Header.Get("x-correlation-id")
	if correlationId == "" {
		return b.GetDefaultCorrelationParams()
	}
	return &log.CorrelationParmas{
		ServiceName:   b.config.GetAppConfig().ServiceName,
		CorrelationId: correlationId,
		ScenarioId:    r.Header.Get("x-scenario-id"),
		ScenarioName:  r.Header.Get("x-scenario-name"),
		SessionId:     r.Header.Get("x-session-id"),
	}
}

func (b *BaseApp) GetDefaultCorrelationParams() *log.CorrelationParmas {
	return &log.CorrelationParmas{
		ServiceName:   b.config.GetAppConfig().ServiceName,
		CorrelationId: fmt.Sprintf("%v-%v", b.config.GetAppConfig().ServiceName, uuid.New().String()),
	}
}

func (b *BaseApp) PrintHeader(ctx context.Context, h http.Header) {
	popList := make(map[string][]string)
	for _, key := range b.config.GetLoggerConfig().AuthHeaderKeyList {
		val := h.Values(key)
		if len(val) != 0 {
			popList[key] = val
			h.Set(key, "<<reducted>>")
		}
	}
	b.log.Info(ctx, "Request-Header", h)
	for key, value := range popList {
		h.Del(key)
		for _, v := range value {
			h.Add(key, v)
		}
	}
}

func (b *BaseApp) PrintBody(ctx context.Context, body []byte) {
	bodyMap := make(map[string]interface{})
	_ = json.Unmarshal(body, &bodyMap)
	b.log.Error(ctx, "Response-Body", bodyMap)
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *BaseApp) GetCorrelationContext(ctx context.Context, c *log.CorrelationParmas) context.Context {
	ctx = context.WithValue(ctx, constant.CorrelationContextKey, c)
	return ctx
}

func (b *BaseApp) GetPort() string {
	return fmt.Sprintf("%v:%v", b.config.GetAppConfig().Host, b.config.GetAppConfig().Port)
}
