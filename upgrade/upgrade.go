package upgrade

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

func startInstance(binPath, tempBind string) error {
	// Potencial security vulnerability; research if f.Name() can be a malicious value.
	cmd := exec.Command(binPath, "-upgrade", "true")
	if err := cmd.Start(); err != nil {
		return err
	}

	zap.L().Debug("start upgraded server", zap.String("bin", binPath), zap.String("bind", tempBind))
	time.Sleep(5 * time.Second)
	return nil
}

func Upgrade(logger *zap.Logger, s *http.Server, binPath, tempBind string) error {
	if err := stopServer(s); err != nil {
		zap.L().Fatal("shutdown server", zap.Error(err))
	}

	if err := startInstance(binPath, tempBind); err != nil {
		return err
	}

	// TODO come on
	_, err := http.Get(tempBind + "replace")
	if err != nil {
		return fmt.Errorf("call /replace: %w", err)
	}
	zap.L().Info("replace successful")

	return nil
}
