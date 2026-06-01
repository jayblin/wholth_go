package container

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"wholth_go/logger"
)

type Container struct {
	Tag string
}

func (c *Container) Log(severity logger.Severity, message string, errs ...error) {
	logger.Log(severity, fmt.Sprintf("[%s]%s", c.Tag, message), errs...)
	// logger.Log(severity, fmt.Sprintf("[%s][%s]", c.Tag, message), errs...)
}

var ContainerKey *Container

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ContainerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cont := Container{
			Tag: randSeq(16),
			// Breadcrumbs: make([]string, 0),
		}
		ctx := context.WithValue(r.Context(), ContainerKey, cont)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Instance(r *http.Request) *Container {
	val, ok := r.Context().Value(ContainerKey).(Container)

	if ok {
		return &val
	}

	return nil

}

func Log(
	r *http.Request,
	severity logger.Severity,
	message string,
	errs ...error,
) {
	instance := Instance(r)

	if nil == instance {
		logger.Log(
			logger.EMERGENCY,
			fmt.Sprintf("[ACCESS_TO_NULL_CONTAINER_INTANCE]%s", message),
			errs...,
		)
	} else {
		logger.Log(
			severity,
			fmt.Sprintf(
				"[%s]%s{\"request_path\":\"%s\",\"request_query\":\"%s\"}",
				instance.Tag,
				message,
				r.URL.Path,
				r.URL.RawQuery,
			),
			errs...,
		)
	}
}

func StatusCodeFromSeverity(sev logger.Severity) int {
	switch sev {
	case logger.EMERGENCY:
		return http.StatusServiceUnavailable
	case logger.ALERT:
		return http.StatusInternalServerError
	case logger.CRITICAL:
		return http.StatusInternalServerError
	case logger.ERROR:
		return http.StatusInternalServerError
	case logger.WARNING:
		return http.StatusInternalServerError
	case logger.NOTICE:
		return http.StatusBadRequest
	case logger.INFO:
		return http.StatusOK
	case logger.DEBUG:
		return http.StatusOK
	}

	return http.StatusOK
}
