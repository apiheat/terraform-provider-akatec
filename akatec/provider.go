package akatec

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"edgerc":           &schema.Schema{Type: schema.TypeString, Required: true},
			"section":          &schema.Schema{Type: schema.TypeString, Required: true},
			"accountSwitchKey": &schema.Schema{Type: schema.TypeString, Optional: true}},

		ResourcesMap: map[string]*schema.Resource{
			// "akatec_network_list": resourceNetlist()
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ProviderMetaSchema: map[string]*schema.Schema{},

		ConfigureFunc: providerConfigure,

		TerraformVersion: "",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		edgerc:           d.Get("edgerc").(string),
		section:          d.Get("section").(string),
		accountSwitchKey: d.Get("accountSwitchKey").(string),
	}

	return config.Client()
}
