package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/service"
	"github.com/alecthomas/kingpin"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
)

func Run(args []string) bool {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	cfg := config.New(kv.MustFromEnv())
	log = cfg.Log()

	app := kingpin.New("uniswapv2-indexer", "")

	runCmd := app.Command("run", "run command")
	serviceCmd := runCmd.Command("service", "run service") // you can insert custom help

	// custom commands go here...

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		log.Info("received interrupt signal, shutting down")
		cancel()
	}()

	switch cmd {
	case serviceCmd.FullCommand():
		service.Run(ctx, cfg)
	// handle any custom commands here in the same way
	default:
		log.Errorf("unknown command %s", cmd)
		return false
	}

	return true
}
