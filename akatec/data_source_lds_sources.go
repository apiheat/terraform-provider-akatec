package akatec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLdsSources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsSourcesRead,
		Schema: map[string]*schema.Schema{
			"sources": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of logdelivery service sources",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_id":        {Type: schema.TypeString, Computed: true},
						"source_type_name": {Type: schema.TypeString, Computed: true},
						"source_cpcode":    {Type: schema.TypeString, Computed: true},
						"source_products":  {Type: schema.TypeList, Computed: true},
					},
				},
			},
		},
	}
}

func dataLdsSourcesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	sources, err := api.ldsv3.ListSources()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, src := range *sources {
		if err := d.Set("source_id", src.ID); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(src.ID)
		if err := d.Set("source_type_name", src.Type); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("source_cpcode", src.CpCode); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("source_products", src.Products); err != nil {
			return diag.FromErr(err)
		}

	}

	return diags
}
