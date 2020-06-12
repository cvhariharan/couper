package config

import (
	"github.com/hashicorp/hcl/v2"
)

const (
	ServeDir  = "ServeDir"
	ServeFile = "ServeFile"
)

type Backend struct {
	Kind        string   `hcl:"kind,label"`
	Name        string   `hcl:"name,label"`
	Description string   `hcl:"description,optional"`
	Options     hcl.Body `hcl:",remain"`
}
