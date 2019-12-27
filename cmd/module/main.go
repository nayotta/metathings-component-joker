package main

import (
	component "github.com/nayotta/metathings/pkg/component"

	service "github.com/nayotta/metathings-component-joker/pkg/joker/service"
)

func main() {
	mdl, err := component.NewModule("joker", new(service.JokerService))
	if err != nil {
		panic(err)
	}

	if err = mdl.Launch(); err != nil {
		panic(err)
	}
}
