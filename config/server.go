package config

import "net/http"

type Server struct {
	BasePath    string   `hcl:"base_path,optional"`
	Name        string   `hcl:"name,label"`
	Domains     []string `hcl:"domains,optional"`
	Files       *Files   `hcl:"files,block"`
	Spa         *Spa     `hcl:"spa,block"`
	Api         *Api     `hcl:"api,block"`
	FileHandler http.Handler
}
