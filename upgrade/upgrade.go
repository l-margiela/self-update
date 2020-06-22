package upgrade

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

var errInvalidBind = errors.New("invalid bind")

// urlify returns bind string (e.g. ":8080") formatted as a proper URL.
func urlify(bind string) (url.URL, error) {
	split := strings.Split(bind, ":")

	switch {
	case len(split) == 1 || split[0] == "":
		return url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%s", split[1]),
		}, nil
	case len(split) == 2 && split[0] != "":
		return url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("http://%s/", bind),
		}, nil
	default:
		return url.URL{}, errInvalidBind
	}
}

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

func Upgrade(logger *zap.Logger, s *http.Server, binPath, tempBind, bind string) error {
	if err := stopServer(s); err != nil {
		zap.L().Fatal("shutdown server", zap.Error(err))
	}

	if err := startInstance(binPath, tempBind, bind); err != nil {
		return err
	}

	url, err := urlify(tempBind)
	if err != nil {
		return fmt.Errorf(`invalid bind "%s": %w`, tempBind, err)
	}

	url.Path = "/replace"

	resp, err := http.Get(url.String())
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("close response body", zap.Error(err))
		}
	}()

	if err != nil {
		return fmt.Errorf("call %s: %w", url.Path, err)
	}

	zap.L().Info("replace successful")

	return nil
}
