package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Masterminds/semver"
	"go.uber.org/zap"
)

// Version is injected by ld flags. See Makefile.
//
// The value set here is a placeholder used in case
// of invalid build process.
//
// The program will panic on a version invalid with semver.
// See https://semver.org/.
var Version = "unknown"

func startUpgradeServer(server *http.Server, upgradeBind string) {
	tempRouter := http.NewServeMux()
	tempServer := &http.Server{
		Addr:    upgradeBind,
		Handler: tempRouter,
	}

	tempRouter.HandleFunc("/replace", replaceHandler(tempServer, server))

	go func() {
		zap.L().Info("start", zap.String("bind", upgradeBind), zap.String("version", Version))
		if err := tempServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				zap.L().Fatal("listen and serve temporary", zap.Error(err))
			}
		}
	}()
}

func setupLogger(dev bool) (func(), func()) {
	var err error

	var logger *zap.Logger
	if dev {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		fmt.Printf(`{"error": "%s"}\n`, err)
		os.Exit(1)
	}

	return func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf(`{"error": "%s"}\n`, err)
		}
	}, zap.ReplaceGlobals(logger)
}

func main() {
	bind := flag.String("bind", ":8080", "Host and port pair")
	version := flag.Bool("version", false, "Display version")
	upgradeBind := flag.String("upgrade-bind", ":8081", "Defines temporary port used during upgrade process")
	upgradeMode := flag.Bool("upgrade", false, "Used by the upgrade mechanism")
	dev := flag.Bool("dev", false, "Development mode")
	flag.Parse()

	sync, undo := setupLogger(*dev)
	defer sync()
	defer undo()

	_, err := semver.NewVersion(Version)
	if err != nil {
		zap.L().Error("parse version", zap.Error(err))
	}

	if *version {
		fmt.Println(Version)
		return
	}

	router := http.NewServeMux()
	server := &http.Server{
		Addr:    *bind,
		Handler: router,
	}

	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/check", checkHandler)
	router.HandleFunc("/upgrade", upgradeHandler(server, *upgradeBind))

	if *upgradeMode {
		startUpgradeServer(server, *upgradeBind)
	} else {
		go func() {
			zap.L().Info("start", zap.String("bind", *bind), zap.String("version", Version))
			if err := server.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					zap.L().Fatal("listen and serve", zap.Error(err))
				}
			}
		}()
	}

	// FIXME: graceful exit on signal

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	zap.L().Info("catch signal", zap.Stringer("signal", sig))
}
