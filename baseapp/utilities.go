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
	"sabariram.com/goserverbase/utils"
)

func (b *BaseApp) GetHttpCorrelationParams(r *http.Request) *log.CorrelationParmas {
	correlationId := r.Header.Get("x-correlation-id")
	if correlationId == "" {
		return log.GetDefaultCorrelationParams(b.c.AppConfig.ServiceName)
	}
	return &log.CorrelationParmas{
		ServiceName:   b.c.AppConfig.ServiceName,
		CorrelationId: correlationId,
		ScenarioId:    r.Header.Get("x-scenario-id"),
		ScenarioName:  r.Header.Get("x-scenario-name"),
		SessionId:     r.Header.Get("x-session-id"),
	}
}

func (b *BaseApp) PrintRequest(ctx context.Context, r *http.Request) {
	h := r.Header
	popList := make(map[string][]string)
	for _, key := range b.c.LoggerConfig.AuthHeaderKeyList {
		val := h.Values(key)
		if len(val) != 0 {
			popList[key] = val
			h.Set(key, "---reducted---")
		}
	}
	b.log.Info(ctx, "Request", map[string]interface{}{
		"Method":        r.Method,
		"Header":        h,
		"URL":           r.URL,
		"Proto":         r.Proto,
		"ContentLength": r.ContentLength,
		"Host":          r.Host,
		"RemoteAddr":    r.RemoteAddr,
		"RequestURI":    r.RequestURI,
	})
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
		return nil, fmt.Errorf("baseapp.GetBytes: %w", err)
	}
	return buf.Bytes(), nil
}

func (b *BaseApp) GetCorrelationContext(ctx context.Context, c *log.CorrelationParmas) context.Context {
	ctx = context.WithValue(ctx, constant.CorrelationContextKey, c)
	return ctx
}

func (b *BaseApp) GetPort() string {
	return fmt.Sprintf("%v:%v", b.c.AppConfig.Host, b.c.AppConfig.Port)
}

type Filter struct {
	PageNo int64  `json:"pageNo" schema:"pageNo"`
	Limit  int64  `json:"limit" schema:"limit"`
	SortBy string `json:"sortby" schema:"sortby"`
	Asc    *bool  `json:"asc" schema:"asc"`
}

func SetDefaultPagination(filter interface{}, deafultSortBy string) error {
	var defaultFilter Filter
	err := utils.JsonTransformer(filter, &defaultFilter)
	if err != nil {
		return fmt.Errorf("baseapp.SetDefault : %w", err)
	}
	if defaultFilter.PageNo <= 0 {
		defaultFilter.PageNo = 1
	}
	if defaultFilter.Limit <= 0 {
		defaultFilter.Limit = 10
	}
	if defaultFilter.SortBy == "" {
		defaultFilter.SortBy = deafultSortBy
	}
	if defaultFilter.Asc == nil {
		v := true
		defaultFilter.Asc = &v
	}
	err = utils.JsonTransformer(&defaultFilter, filter)
	if err != nil {
		return fmt.Errorf("app.SetDefault : %w", err)
	}
	return nil
}
