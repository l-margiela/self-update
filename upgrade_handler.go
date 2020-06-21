package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/xaxes/self-update/check"
	"github.com/xaxes/self-update/upgrade"
	"go.uber.org/zap"
)

func newestCandidateErr(err error, w http.ResponseWriter) {
	zap.L().Error("get newest upgrade candidate", zap.Error(err))
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
}

func upgradeHandler(s *http.Server, tempBind string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		c, err := check.NewestCandidate(check.UpdateBinPrefix, "./", Version)
		if err != nil {
			newestCandidateErr(err, w)
			return
		}

		go func() {
			if err := upgrade.Upgrade(zap.L(), s, "./"+c.Name(), tempBind); err != nil {
				zap.L().Error("upgrade", zap.Error(err), zap.String("status", "failure"))
				os.Exit(1)
			}
			zap.L().Info("upgrade", zap.String("status", "success"))
			os.Exit(0)
		}()
	}
}
