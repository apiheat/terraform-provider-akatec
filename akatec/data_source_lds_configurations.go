package akatec

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceLdsConfigurations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsConfigurationsRead,
		Description: descriptions["configurations"],
		Schema: map[string]*schema.Schema{
			"source_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["source_type"],
				ValidateFunc: validation.StringInSlice([]string{
					"answerx",
					"cpcode-products",
					"edns",
					"gtm",
				}, false),
			},
			"configurations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: descriptions["configuration"],
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
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
		item := map[string]interface{}{
			"id":              strconv.Itoa(cfg.ID),
			"status":          cfg.Status,
			"start_date":      cfg.StartDate,
			"log_source_id":   cfg.LogSource.ID,
			"log_source_type": cfg.LogSource.Type,
		}
		if cfg.EndDate != "" {
			item["end_date"] = cfg.EndDate
		}
		switch cfg.LogSource.Type {
		case "cpcode-products":
			item["log_source_cpcode"] = cfg.LogSource.CpCode
		case "answerx":
			item["log_source_name"] = cfg.LogSource.Name
		case "edns":
			item["log_source_zone_name"] = cfg.LogSource.ZoneName
		case "gtm":
			item["log_source_property_name"] = cfg.LogSource.PropertyName
		}
		cfgsList = append(cfgsList, item)
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
