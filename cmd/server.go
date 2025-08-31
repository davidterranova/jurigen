package cmd

import (
	"context"
	"davidterranova/jurigen/pkg"
	"davidterranova/jurigen/pkg/port"
	"davidterranova/jurigen/pkg/xhttp"
	"os"
	"os/signal"
	"syscall"

	ihttp "davidterranova/jurigen/pkg/adapter/http"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts jurigen server",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := pkg.New(port.NewFileDAGRepository("./data"))

	go httpAPIServer(ctx, app)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		cancel()
	case <-ctx.Done():
	}
}

func httpAPIServer(ctx context.Context, app *pkg.App) {
	router := ihttp.New(
		app,
		nil,
		// xhttp.GrantAnyFn(),
	)
	server := xhttp.NewServer(router, "", 8080)

	err := server.Serve(ctx)
	if err != nil {
		log.Ctx(ctx).Panic().Err(err).Msg("failed to start http server")
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
