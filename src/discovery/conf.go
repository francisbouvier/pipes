package discovery

import (
	"encoding/json"
	"errors"

	"github.com/codegangsta/cli"
)

type Conf struct {
	path      string
	discovery Discovery
	data      map[string]interface{}
}

func (c *Conf) GetPool(name string) (p string, err error) {
	pools := c.data["pools"].(map[string]interface{})
	po, prs := pools[name]
	if prs == false {
		err = errors.New("Pool does not exist")
		return
	}
	p = po.(string)
	return
}

func (c *Conf) SetPool(name, value string) {
	pools := c.data["pools"].(map[string]interface{})
	pools[name] = value
}

func (c *Conf) DeletePool(name string) {
	pools := c.data["pools"].(map[string]interface{})
	delete(pools, name)
}

func (c *Conf) GetMainPool() (name string, value string, err error) {
	name = c.data["main_pool"].(string)
	value, err = c.GetPool(name)
	return
}

func (c *Conf) SetMainPool(name string) {
	c.data["main_pool"] = name
}

func (c *Conf) Save() error {
	data, err := json.MarshalIndent(c.data, "", "    ")
	if err != nil {
		return err
	}
	return c.discovery.Save(c.path, data)
}

func getConf(c *cli.Context) (cf *Conf, err error) {
	if err != nil {
		return
	}
	cf = &Conf{path: "discovery.json"}
	var d Discovery
	switch c.String("service") {
	default:
		d = Local{}
		cf.discovery = d
	}
	dataB, err := d.Connect(cf.path)
	data := cf.data
	err = json.Unmarshal(dataB, &data)
	if err == nil {
		cf.data = data

	} else {
		cf.data = map[string]interface{}{
			"main_pool": "",
			"pools":     map[string]interface{}{},
		}
		err = nil
	}
	return
}
