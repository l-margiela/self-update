package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

func startNewer(c os.FileInfo, tempBind string) error {
	// Potencial security vulnerability; research if f.Name() can be a malicious value.
	cmd := exec.Command("./"+c.Name(), "-upgrade", "true")
	if err := cmd.Start(); err != nil {
		return err
	}

	zap.L().Debug("start upgraded server", zap.String("bin", c.Name()), zap.String("bind", tempBind))
	time.Sleep(5 * time.Second)
	return nil
}

func upgradeHandler(s *http.Server, tempBind string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		c, err := NewestCandidate(updateBinPrefix, "./")
		if err != nil {
			zap.L().Error("get newest upgrade candidate", zap.Error(err))
			if errors.Is(err, ErrNoCandidate) {
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

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			zap.L().Fatal("shutdown server", zap.Error(err))
		}

		// Down from here, the server is shut

		go func(s *http.Server, tempBind string, c os.FileInfo) {
			if err := startNewer(c, tempBind); err != nil {
				zap.L().Error("handle UNIX", zap.Error(err))
				return
			}

			u, err := urlify(tempBind)
			if err != nil {
				zap.L().Fatal("parse bind", zap.String("bind", tempBind))
			}
			_, err = http.Get(u + "replace")
			if err != nil {
				zap.L().Fatal("call /replace", zap.Error(err))
				return
			}
			zap.L().Info("replace successful")
			os.Exit(0)
		}(s, tempBind, c)
	}
}
