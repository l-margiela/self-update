package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xaxes/self-update/check"
	"go.uber.org/zap"
)

func checkHandler(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
	new, err := check.NewestCandidate(Version, "./", Version)
	if err != nil {
		if errors.Is(err, check.ErrNoCandidate) {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				zap.L().Error("write response", zap.Error(err))
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			zap.L().Error("write response", zap.Error(err))
		}
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("cadidate: %v", new.Name()))); err != nil {
		zap.L().Error("write response", zap.Error(err))
	}
}
