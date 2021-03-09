package akatec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLdsDeliveryThresholds() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsDeliveryThresholdsRead,
		Description: descriptions["delivery_thresholds"],
		Schema: map[string]*schema.Schema{
			"thresholds": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: descriptions["delivery_threshold"],
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":   {Type: schema.TypeString, Computed: true},
						"name": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceLdsDeliveryThreshold() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsDeliveryThresholdRead,
		Description: descriptions["delivery_threshold"],
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataLdsDeliveryThresholdsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	data, err := api.ldsv3.ListDeliveryThresholds()
	if err != nil {
		return diag.FromErr(err)
	}

	list := make([]map[string]interface{}, 0, len(*data))

	for _, item := range *data {
		list = append(list, flattenLdsParameter(&item))
	}

	if err := d.Set("thresholds", list); err != nil {
		return diag.FromErr(fmt.Errorf("%q", err.Error()))
	}

	jsonBody, err := json.Marshal(list)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSHAString(string(jsonBody)))

	return diags
}

func dataLdsDeliveryThresholdRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	fName := d.Get("name").(string)

	data, err := api.ldsv3.ListDeliveryThresholds()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, item := range *data {
		if fName == item.Value {
			d.SetId(item.ID)
			return diags
		}
	}

	return diags
}
