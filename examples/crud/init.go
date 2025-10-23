package main

import "github.com/dracory/liveflux"

func init() {
    liveflux.Register(new(UserList))
    liveflux.Register(new(CreateUserModal))
    liveflux.Register(new(EditUserModal))
    liveflux.Register(new(DeleteUserModal))
}
