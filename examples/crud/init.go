package main

import (
	"log"

	"github.com/dracory/liveflux"
)

func init() {
	if err := liveflux.Register(new(UserList)); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.Register(new(CreateUserModal)); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.Register(new(EditUserModal)); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.Register(new(DeleteUserModal)); err != nil {
		log.Fatal(err)
	}
}
