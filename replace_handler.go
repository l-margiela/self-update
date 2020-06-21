package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func replaceHandler(tempServer, server *http.Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		// the goroutine here is needed because the code below closes
		// the server, so we wouldn't be able to respond to the request
		// properly.
		go func(tempServer, server *http.Server) {
			// It can be done in a better style using a custom HTTP server.
			time.Sleep(1 * time.Second)
			// we don't care if it closes the connections successfully, hence 0
			ctx, cancel := context.WithTimeout(context.Background(), 10)
			defer cancel()
			if err := tempServer.Shutdown(ctx); err != nil {
				zap.L().Error("shutdown temporary server", zap.Error(err))
			}

			go func() {
				if err := server.ListenAndServe(); err != nil {
					if !errors.Is(err, http.ErrServerClosed) {
						zap.L().Fatal("listen and serve on replace", zap.Error(err))
					}
				}
			}()
		}(tempServer, server)
	}
}
