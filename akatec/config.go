package akatec

import (
	edgegrid "github.com/apiheat/go-edgegrid/v6/edgegrid"
	svcNetList "github.com/apiheat/go-edgegrid/v6/service/netlistv2"
)

type AkamaiServices struct {
	netlistV2 *svcNetList.Netlistv2
}

type Config struct {
	edgerc           string
	section          string
	accountSwitchKey string
}

func (c *Config) Client() (*AkamaiServices, error) {

	var creds *edgegrid.Credentials

	clientSvc := AkamaiServices{
		netlistV2: &svcNetList.Netlistv2{},
	}
	var err error

	creds, err = edgegrid.NewCredentials().FromFile(c.edgerc).Section(c.section)

	if err != nil {
		return nil, err
	}

	config := edgegrid.NewConfig().
		WithCredentials(creds)

	if c.accountSwitchKey != "" {
		config = config.WithAccountSwitchKey(c.accountSwitchKey)
	}

	clientSvc.netlistV2 = svcNetList.New(config)

	return &clientSvc, nil
}
