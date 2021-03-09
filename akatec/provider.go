package akatec

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	edgegrid "github.com/apiheat/go-edgegrid/v6/edgegrid"
	akalds "github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	akanetlist "github.com/apiheat/go-edgegrid/v6/service/netlistv2"
)

var descriptions map[string]string

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"edgerc": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
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
			"akatec_lds_delivery_thresholds":  dataSourceLdsDeliveryThresholds(),
			"akatec_lds_delivery_threshold":   dataSourceLdsDeliveryThreshold(),
			"akatec_lds_contacts":             dataSourceLdsContacts(),
			"akatec_lds_contact":              dataSourceLdsContact(),
			"akatec_lds_netstorage_groups":    dataSourceLdsNetStorageGroups(),
			"akatec_lds_netstorage_group":     dataSourceLdsNetStorageGroup(),
			"akatec_lds_message_sizes":        dataSourceLdsMessageSizes(),
			"akatec_lds_message_size":         dataSourceLdsMessageSize(),
			"akatec_lds_encodings":            dataSourceLdsEncodings(),
			"akatec_lds_encoding":             dataSourceLdsEncoding(),
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

	edgerc := d.Get("edgerc").(string)
	if edgerc == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "NewCredentials | Unable to create Akamai client",
				Detail:   "NewCredentials | Unable to create credentials edgerc file path cannot be found",
			})
			return nil, diags
		}
		edgerc = fmt.Sprintf("%s/.edgerc", homeDir)
	}

	creds, err := edgegrid.NewCredentials().FromFile(edgerc).Section(d.Get("section").(string))
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

func init() {
	descriptions = map[string]string{
		"source_type":    "Type of log source for which you can create/request a log delivery configuration.",
		"contacts":       "List of logdelivery service available contacts",
		"contact":        "Contact information",
		"configurations": "List of log delivery service available configurations",
		"configuration": "You can create one LDS configuration for each log source.\n" +
			"Once you initially configure an Akamai service for log delivery, you can view, edit,\n" +
			"and suspend and reactivate that log delivery configuration.",
		"delivery_frequency":   "Used with aggregation type byLogArrival. Period of time that will be covered by log delivery.",
		"delivery_frequencies": "List of log delivery service available delivery frequencies",
		"delivery_threshold": "Used with aggregation type byHitTime. Data completion threshold,\n" +
			"or the percentage of expected logs to be processed before the log data is sent to you",
		"delivery_thresholds": "List of log delivery service available delivery thresholds",
		"encoding":            "Selected encoding option used to encode logs",
		"encodings":           "List of log delivery service available encoding options",
		"delivery_type":       "Encapsulates log delivery via supported types",
		"log_format":          "Selected format for log delivery",
		"log_formats":         "List of log delivery service available log formats",
		"message_size":        "Packed log message’s approximate size",
		"message_sizes":       "List of log delivery service available message sizes",
		"example":             "",
	}
}
