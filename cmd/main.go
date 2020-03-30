package main

import (
	"log"

	"golang.org/x/net/context"

	"github.com/luigi-riefolo/xyz/api"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ns, err := api.NewXYZService(ctx)
	if err != nil {
		log.Fatalf("could not create api: %#v", err)
	}

	s, _ := ns.(*api.Service)
	if err := s.Start(ctx); err != nil {
		log.Fatalf("could not start server: %#v", err)
	}
}
