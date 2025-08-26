package util

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GetCurrentTime возвращает текущее время в формате RFC3339
func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

// UpdateRoutesReq представляет запрос на обновление маршрутов сервиса.
type UpdateRoutesReq struct {
	ServiceName string `json:"service_name"` // Имя сервиса
	URL         string `json:"URL"`          // URL сервиса
}

type route struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

// GetMetricsRoute добавляет маршрут для метрик Prometheus (/metrics).
// Принимает gin.Engine для регистрации маршрута.
func GetMetricsRoute(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// GetHealthcheckRoute добавляет маршрут для проверки работоспособности (/healthcheck).
// Принимает gin.Engine для регистрации маршрута.
func GetHealthcheckRoute(r *gin.Engine) {
	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}

// GetRouteList добавляет маршрут для получения списка всех маршрутов (/routes).
// Принимает gin.Engine для регистрации маршрута.
func GetRouteList(r *gin.Engine) {
	r.GET("/routes", func(c *gin.Context) {
		resp := make([]route, 0)
		for _, r := range r.Routes() {
			resp = append(resp, route{
				Method:  r.Method,
				Path:    r.Path,
				Handler: r.Handler,
			})
		}

		c.JSON(http.StatusOK, resp)
	})
}
