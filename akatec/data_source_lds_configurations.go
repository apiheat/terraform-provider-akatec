package akatec

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceLdsConfigurations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsConfigurationsRead,
		Schema: map[string]*schema.Schema{
			"source_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"answerx",
					"cpcode-products",
					"edns",
					"gtm",
				}, false),
			},
			"configurations": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of logdelivery service configurations",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                           {Type: schema.TypeInt, Computed: true},
						"status":                       {Type: schema.TypeString, Computed: true},
						"start_date":                   {Type: schema.TypeString, Computed: true},
						"end_date":                     {Type: schema.TypeString, Computed: true},
						"associated_log_source_id":     {Type: schema.TypeString, Computed: true},
						"associated_log_source_cpcode": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataLdsConfigurationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	srcType, srcTypeOk := d.GetOk("source_type")
	if !srcTypeOk {
		return diag.Errorf("Type is not assigned. Please provide correct type")
	}

	cfgs, err := api.ldsv3.ListLogConfigurationsByType(srcType.(string))
	if err != nil {
		return diag.FromErr(err)
	}

	cfgsList := make([]map[string]interface{}, 0, len(*cfgs))

	for _, cfg := range *cfgs {
		cfgsList = append(cfgsList, map[string]interface{}{
			"id":         cfg.ID,
			"status":     cfg.Status,
			"start_date": cfg.StartDate,
			// Add once client will support that
			"end_date":                     "N/A",
			"associated_log_source_id":     cfg.LogSource.ID,
			"associated_log_source_cpcode": cfg.LogSource.CpCode,
		})
	}

	if err := d.Set("configurations", cfgsList); err != nil {
		return diag.FromErr(err)
	}

	jsonBody, err := json.Marshal(cfgsList)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSHAString(string(jsonBody)))

	return diags
}
