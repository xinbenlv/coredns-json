package json

import (
	"net/http"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	caddy.RegisterPlugin("json", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	j, err := parseJSON(c)
	if err != nil {
		return plugin.Error("json", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		j.Next = next
		return j
	})

	return nil
}

func parseJSON(c *caddy.Controller) (*JSON, error) {
	j := &JSON{
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		DNSSEC: false,
	}

	for c.Next() {
		args := c.RemainingArgs()
		if len(args) != 1 {
			return nil, c.ArgErr()
		}
		j.URL = args[0]

		for c.NextBlock() {
			switch c.Val() {
			case "dnssec":
				j.DNSSEC = true
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}

	return j, nil
}
