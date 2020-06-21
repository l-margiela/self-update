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

var page = `<!DOCTYPE html>
<html>
	<head>
		<title>Upgrade</title>
	</head>
	<body>
		<h1>Redirecting...</h1>
		<script type="text/javascript">
			setTimeout(()=>{
				window.location.href = "/";
			}, 10000)
		</script>
	</body>
</html>
`

func upgradeHandler(s *http.Server, tempBind string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))

		c, err := check.NewestCandidate(".", Version)
		if err != nil {
			newestCandidateErr(err, w)
			return
		}

		if _, err := w.Write([]byte(page)); err != nil {
			zap.L().Error("write response", zap.Error(err))
		}

		go func() {
			if err := upgrade.Upgrade(zap.L(), s, c.Path, tempBind); err != nil {
				zap.L().Fatal("upgrade", zap.Error(err), zap.String("status", "failure"))
			}

			zap.L().Info("upgrade", zap.String("status", "success"))
			os.Exit(0)
		}()
	}
}
