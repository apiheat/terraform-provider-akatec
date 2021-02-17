package akatec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	edgegrid "github.com/apiheat/go-edgegrid/v6/edgegrid"
	akalds "github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	akanetlist "github.com/apiheat/go-edgegrid/v6/service/netlistv2"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"edgerc":  {Type: schema.TypeString, Required: true},
			"section": {Type: schema.TypeString, Required: true},
			"ask":     {Type: schema.TypeString, Optional: true}},

		ResourcesMap: map[string]*schema.Resource{
			"akatec_netlist_ip":        resourceNetlistIP(),
			"akatec_netlist_geo":       resourceNetlistGeo(),
			"akatec_lds_configuration": resourceLdsConfiguration(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"akatec_netlist_ip":               dataSourceNetlistIP(),
			"akatec_lds_sources":              dataSourceLdsSources(),
			"akatec_lds_configurations":       dataSourceLdsConfigurations(),
			"akatec_lds_log_formats":          dataSourceLdsLogFormats(),
			"akatec_lds_log_format":           dataSourceLdsLogFormat(),
			"akatec_lds_delivery_frequencies": dataSourceLdsDeliveryFrequencies(),
			"akatec_lds_delivery_frequency":   dataSourceLdsDeliveryFrequency(),
		},

		ProviderMetaSchema:   map[string]*schema.Schema{},
		ConfigureContextFunc: providerConfigure,
		TerraformVersion:     "",
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var (
		creds *edgegrid.Credentials
		diags diag.Diagnostics
	)

	apiClient := AkamaiServices{
		netlistV2: &akanetlist.Netlistv2{},
		ldsv3:     &akalds.Ldsv3{},
	}

	creds, err := edgegrid.NewCredentials().FromFile(d.Get("edgerc").(string)).Section(d.Get("section").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "NewCredentials | Unable to create Akamai client",
			Detail:   "NewCredentials | Unable to create credentials based on edgerc file and section",
		})
		return nil, diags
	}

	config := edgegrid.NewConfig().
		WithCredentials(creds).WithLogVerbosity("Debug").WithRequestDebug(true)

	ask, askOk := d.GetOk("new_attribute")
	if askOk {
		config = config.WithAccountSwitchKey(ask.(string))
	}

	apiClient.netlistV2 = akanetlist.New(config)
	apiClient.ldsv3 = akalds.New(config)

	return &apiClient, diags

}
