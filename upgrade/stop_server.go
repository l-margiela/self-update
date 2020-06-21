package upgrade

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func stopServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return nil
}
