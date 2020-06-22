package upgrade

import (
	"os/exec"
	"time"

	"go.uber.org/zap"
)

func startInstance(binPath, tempBind, bind string) error {
	// FIXME: Potential security vulnerability; research if binPath can be a malicious value.
	cmd := exec.Command(binPath, "-upgrade", "true", "-upgrade-bind", tempBind, "-bind", bind)
	if err := cmd.Start(); err != nil {
		return err
	}

	zap.L().Debug("start upgraded server", zap.String("bin", binPath), zap.String("bind", tempBind))
	time.Sleep(5 * time.Second)
	return nil
}
