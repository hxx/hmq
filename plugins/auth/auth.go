package auth

import (
	authfile "github.com/fhmq/hmq/plugins/auth/authfile"
	"github.com/fhmq/hmq/plugins/auth/authhttp"
	"github.com/fhmq/hmq/plugins/auth/authmango"
)

const (
	AuthMango = "authmango"
	AuthHTTP  = "authhttp"
	AuthFile  = "authfile"
)

type Auth interface {
	CheckACL(action, clientID, username, ip, topic string) bool
	CheckConnect(clientID, username, password string) bool
}

func NewAuth(name string) Auth {
	switch name {
	case AuthMango:
		return authmango.Init()
	case AuthHTTP:
		return authhttp.Init()
	case AuthFile:
		return authfile.Init()
	default:
		return &mockAuth{}
	}
}
