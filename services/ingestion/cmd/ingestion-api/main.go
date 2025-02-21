package main

import (
	"context"
	"log"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.Provide(
			func() (*zap.Logger, error) {
				return zap.NewProduction()
			},
		),

		fx.Invoke(func(*zap.Logger) {}),
	)

	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	<-app.Done()

	if err := app.Stop(context.Background()); err != nil {
		log.Fatal(err)
	}
}
