package config

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/pipehub/pipehub/internal/application/generator"
)

// Config has the configuration needed by PipeHub.
type Config struct {
	HTTP []configHTTP
	Pipe []configPipe
	Core []configCore `mapstructure:"core"`
}

// ToGenerator generate the config struct needed to run the generator application.
func (c Config) ToGenerator() generator.ClientConfig {
	var cfg generator.ClientConfig
	for _, pipe := range c.Pipe {
		cfg.Pipes = append(cfg.Pipes, generator.Pipe{
			Alias:      pipe.Alias,
			ImportPath: pipe.Path,
			Module:     pipe.Module,
			Version:    pipe.Version,
		})
	}
	return cfg
}

func (c *Config) init(payload []byte) error {
	// For some reason I can't unmarshal direct from the HCL to a struct, the array values get messed up.
	// Unmarshalling to a map works fine, so we do this and later transform the map into the desired struct.
	rawCfg := make(map[string]interface{})
	if err := hcl.Unmarshal(payload, &rawCfg); err != nil {
		return errors.Wrap(err, "unmarshal payload error")
	}

	if err := mapstructure.Decode(rawCfg, c); err != nil {
		return errors.Wrap(err, "unmarshal error")
	}

	var err error
	c.HTTP, err = loadConfigHTTP(rawCfg["http"])
	if err != nil {
		return errors.Wrap(err, "unmarshal http config error")
	}

	c.Pipe, err = loadConfigPipe(rawCfg["pipe"])
	if err != nil {
		return errors.Wrap(err, "unmarshal pipe config error")
	}

	return nil
}

// NewConfig return a configured config.
func NewConfig(payload []byte) (Config, error) {
	var c Config
	if err := c.init(payload); err != nil {
		return c, errors.Wrap(err, "initialization error")
	}
	return c, nil
}

func (c Config) valid() error {
	if len(c.Core) > 1 {
		return errors.New("more then one 'core' config block found, only one is allowed")
	}

	for _, core := range c.Core {
		if err := core.valid(); err != nil {
			return err
		}
	}
	return nil
}

// func (c Config) toClientConfig() (pipehub.ClientConfig, error) {
// 	cfg := pipehub.ClientConfig{
// 		AsyncErrHandler: asyncErrHandler,
// 		HTTP:            make([]pipehub.ClientConfigHTTP, 0, len(c.HTTP)),
// 	}

// 	for _, http := range c.HTTP {
// 		cfg.HTTP = append(cfg.HTTP, pipehub.ClientConfigHTTP{
// 			Endpoint: http.Endpoint,
// 			Handler:  http.Handler,
// 		})
// 	}

// 	for _, pipe := range c.Pipe {
// 		cfg.Pipe = append(cfg.Pipe, pipehub.ClientConfigPipe{
// 			Path:    pipe.Path,
// 			Module:  pipe.Module,
// 			Version: pipe.Version,
// 			Config:  pipe.Config,
// 		})
// 	}

// 	if (len(c.Core) > 0) && (len(c.Core[0].HTTP) > 0) && (len(c.Core[0].HTTP[0].Server) > 0) {
// 		if len(c.Core[0].HTTP[0].Server[0].Action) > 0 {
// 			cfg.Core.HTTP.Server.Action.NotFound = c.Core[0].HTTP[0].Server[0].Action[0].NotFound
// 			cfg.Core.HTTP.Server.Action.Panic = c.Core[0].HTTP[0].Server[0].Action[0].Panic
// 		}

// 		if len(c.Core[0].HTTP[0].Server[0].Listen) > 0 {
// 			cfg.Core.HTTP.Server.Listen.Port = c.Core[0].HTTP[0].Server[0].Listen[0].Port
// 		}
// 	}

// 	if (len(c.Core) > 0) && (len(c.Core[0].HTTP) > 0) && (len(c.Core[0].HTTP[0].Client) > 0) {
// 		client := c.Core[0].HTTP[0].Client[0]
// 		cfg.Core.HTTP.Client.DisableCompression = client.DisableCompression
// 		cfg.Core.HTTP.Client.DisableKeepAlive = client.DisableKeepAlive
// 		cfg.Core.HTTP.Client.MaxConnsPerHost = client.MaxConnsPerHost
// 		cfg.Core.HTTP.Client.MaxIdleConns = client.MaxIdleConns
// 		cfg.Core.HTTP.Client.MaxIdleConnsPerHost = client.MaxIdleConnsPerHost

// 		var err error
// 		cfg.Core.HTTP.Client.ExpectContinueTimeout, err = time.ParseDuration(client.ExpectContinueTimeout)
// 		if err != nil {
// 			return pipehub.ClientConfig{}, errors.Wrapf(err, "parse duration '%s' error", client.ExpectContinueTimeout)
// 		}

// 		cfg.Core.HTTP.Client.IdleConnTimeout, err = time.ParseDuration(client.IdleConnTimeout)
// 		if err != nil {
// 			return pipehub.ClientConfig{}, errors.Wrapf(err, "parse duration '%s' error", client.IdleConnTimeout)
// 		}

// 		cfg.Core.HTTP.Client.TLSHandshakeTimeout, err = time.ParseDuration(client.TLSHandshakeTimeout)
// 		if err != nil {
// 			return pipehub.ClientConfig{}, errors.Wrapf(err, "parse duration '%s' error", client.TLSHandshakeTimeout)
// 		}
// 	}

// 	return cfg, nil
// }

func (c Config) ctxShutdown() (context.Context, func(), error) {
	if (len(c.Core) == 0) || (c.Core[0].GracefulShutdown == "") {
		return context.Background(), func() {}, nil
	}

	raw := c.Core[0].GracefulShutdown
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "parse duration '%s' error", raw)
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), duration)
	return ctx, ctxCancel, nil
}

type configPipe struct {
	Path    string
	Version string
	Alias   string
	Module  string
	Config  map[string]interface{}
}

type configHTTP struct {
	Endpoint string
	Handler  string
}

type configCore struct {
	GracefulShutdown string           `mapstructure:"graceful-shutdown"`
	HTTP             []configCoreHTTP `mapstructure:"http"`
}

func (c configCore) valid() error {
	if len(c.HTTP) > 1 {
		return errors.New("more then one 'core.http' config block found, only one is allowed")
	}

	for _, http := range c.HTTP {
		if err := http.valid(); err != nil {
			return errors.Wrap(err, "core.http invalid")
		}
	}

	return nil
}

type configCoreHTTP struct {
	Server []configCoreHTTPServer `mapstructure:"server"`
	Client []configCoreHTTPClient `mapstructure:"client"`
}

func (c configCoreHTTP) valid() error {
	if len(c.Server) > 1 {
		return errors.New("more then one 'core.http.server' config block found, only one is allowed")
	}

	for _, s := range c.Server {
		if err := s.valid(); err != nil {
			return errors.Wrap(err, "invalid 'core.http.server'")
		}
	}

	if len(c.Client) > 1 {
		return errors.New("more then one 'core.http.client' config block found, only one is allowed")
	}

	return nil
}

type configCoreHTTPServer struct {
	Listen []configServerHTTPListen `mapstructure:"listen"`
	Action []configServerHTTPAction `mapstructure:"action"`
}

func (c configCoreHTTPServer) valid() error {
	if len(c.Action) > 1 {
		return errors.New("more then one 'core.server.http.action' config block found, only one is allowed")
	}

	return nil
}

type configCoreHTTPClient struct {
	DisableKeepAlive      bool   `mapstructure:"disable-keep-alive"`
	DisableCompression    bool   `mapstructure:"disable-compression"`
	MaxIdleConns          int    `mapstructure:"max-idle-conns"`
	MaxIdleConnsPerHost   int    `mapstructure:"max-idle-conns-per-host"`
	MaxConnsPerHost       int    `mapstructure:"max-conns-per-host"`
	IdleConnTimeout       string `mapstructure:"idle-conn-timeout"`
	TLSHandshakeTimeout   string `mapstructure:"tls-handshake-timeout"`
	ExpectContinueTimeout string `mapstructure:"expect-continue-timeout"`
}

type configServerHTTPListen struct {
	Port int `mapstructure:"port"`
}

type configServerHTTPAction struct {
	NotFound string `mapstructure:"not-found"`
	Panic    string `mapstructure:"panic"`
}

// loadConfigPipe expect to receive a interface with this format:
//
//	[]map[string]interface {}{
//		{
//				"github.com/pipehub/pipe": []map[string]interface {}{
//						{
//								"version": "v0.7.0",
//								"alias":   "pipe",
//						},
//				},
//		},
//	}
func loadConfigPipe(raw interface{}) ([]configPipe, error) {
	var result []configPipe

	if raw == nil {
		return nil, nil
	}

	rawSliceMap, ok := raw.([]map[string]interface{})
	if !ok {
		return nil, errors.New("can't type assertion value into []map[string]interface{} on the first assignment")
	}

	for _, rawMap := range rawSliceMap {
		for key, rawMapEntry := range rawMap {
			rawSliceMapInner, ok := rawMapEntry.([]map[string]interface{})
			if !ok {
				return nil, errors.New("can't type assertion value into []map[string]interface{} on the second assignment")
			}

			for _, rawSliceMapInnerEntry := range rawSliceMapInner {
				ch := configPipe{
					Path: key,
				}

				for innerKey, innerEntry := range rawSliceMapInnerEntry {
					switch innerKey {
					case "version", "alias", "module":
						value, ok := innerEntry.(string)
						if !ok {
							return nil, errors.New("can't type assertion value into string")
						}

						switch innerKey {
						case "version":
							ch.Version = value
						case "alias":
							ch.Alias = value
						case "module":
							ch.Module = value
						}
					case "config":
						values, ok := innerEntry.([]map[string]interface{})
						if !ok {
							return nil, errors.New("can't type assertion value into map[string]interface{}")
						}
						if len(values) > 1 {
							return nil, errors.New("more then one 'config' found at a pipe, only one is allowed")
						}
						ch.Config = values[0]
					default:
						return nil, fmt.Errorf("unknow pipe key '%s'", innerKey)
					}
				}

				result = append(result, ch)
			}
		}
	}

	return result, nil
}

// loadConfigHTTP expect to receive a interface with this format:
//
// []map[string]interface {}{
// 	{
// 			"google": []map[string]interface {}{
// 					{
// 							"handler": "base.Default",
// 					},
// 			},
// 	},
// }
func loadConfigHTTP(raw interface{}) ([]configHTTP, error) {
	var result []configHTTP

	if raw == nil {
		return nil, nil
	}

	rawSliceMap, ok := raw.([]map[string]interface{})
	if !ok {
		return nil, errors.New("can't type assertion value into []map[string]interface{} on the first assignment")
	}

	for _, rawMap := range rawSliceMap {
		for key, rawMapEntry := range rawMap {
			rawSliceMapInner, ok := rawMapEntry.([]map[string]interface{})
			if !ok {
				return nil, errors.New("can't type assertion value into []map[string]interface{} on the second assignment")
			}

			for _, rawSliceMapInnerEntry := range rawSliceMapInner {
				ch := configHTTP{
					Endpoint: key,
				}

				for innerKey, innerEntry := range rawSliceMapInnerEntry {
					value, ok := innerEntry.(string)
					if !ok {
						return nil, errors.New("can't type assertion value into string")
					}

					switch innerKey {
					case "handler":
						ch.Handler = value
					default:
						return nil, fmt.Errorf("unknow http key '%s'", innerKey)
					}
				}

				result = append(result, ch)
			}
		}
	}

	return result, nil
}
