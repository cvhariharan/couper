package config

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"

	"go.avenga.cloud/couper/gateway/backend"
)

var typeMap = map[string]func(*logrus.Entry, hcl.Body) http.Handler{
	"proxy": backend.NewProxy(),
}

func LoadFile(name string, log *logrus.Entry) *Gateway {
	var config *Gateway
	err := hclsimple.DecodeFile(name, nil, config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	return Load(config, log)
}

func Load(config *Gateway, log *logrus.Entry) *Gateway {
	backends := make(map[string]http.Handler)

	for a, server := range config.Server {
		// create backends
		for _, be := range server.Api.Backend {
			if isKeyword(be.Name) {
				log.Fatalf("be name not allowed, reserved keyword: '%s'", be.Name)
			}
			if _, ok := backends[be.Name]; ok {
				log.Fatalf("be name must be unique: '%s'", be.Name)
			}
			backends[be.Name] = newBackend(be.Kind, be.Options, log)
		}

		server.Api.PathHandler = make(PathHandler)

		// map backends to endpoint
		endpoints := make(map[string]bool)
		for e, endpoint := range server.Api.Endpoint {
			config.Server[a].Api.Endpoint[e].Server = server // assign parent
			if endpoints[endpoint.Pattern] {
				log.Fatal("Duplicate endpoint: ", endpoint.Pattern)
			}

			endpoints[endpoint.Pattern] = true
			if endpoint.Backend != "" {
				if _, ok := backends[endpoint.Backend]; !ok {
					log.Fatalf("be %q not found", endpoint.Backend)
				}
				server.Api.PathHandler[endpoint] = backends[endpoint.Backend]
				continue
			}

			content, leftOver, diags := endpoint.Options.PartialContent(server.Api.Schema(true))
			if diags.HasErrors() {
				log.Fatal(diags.Error())
			}
			endpoint.Options = leftOver

			if content == nil || len(content.Blocks) == 0 {
				log.Fatalf("expected be attribute reference or block for endpoint: %s", endpoint)
			}
			kind := content.Blocks[0].Labels[0]

			server.Api.PathHandler[endpoint] = newBackend(kind, content.Blocks[0].Body, log) // inline be
		}

		// serve files
		if server.Files.DocumentRoot != "" {
			fileHandler, err := backend.NewFile(server.Files.DocumentRoot, log, server.Spa.BootstrapFile, server.Spa.Paths)
			if err != nil {
				log.Fatalf("Failed to load configuration: %s", err)
			}
			config.Server[a].FileHandler = fileHandler
		}
	}

	return config
}

func newBackend(kind string, options hcl.Body, log *logrus.Entry) http.Handler {
	if !isKeyword(kind) {
		log.Fatalf("Invalid backend: %s", kind)
	}
	b := typeMap[strings.ToLower(kind)](log, options)

	return b
}

func isKeyword(other string) bool {
	_, yes := typeMap[other]
	return yes
}
