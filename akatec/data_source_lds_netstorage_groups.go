package akatec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLdsNetStorageGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsNetStorageGroupsRead,
		Schema: map[string]*schema.Schema{
			"netstorage_groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of logdelivery service sources",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":            {Type: schema.TypeString, Computed: true},
						"name":          {Type: schema.TypeString, Computed: true},
						"cp_code":       {Type: schema.TypeString, Computed: true},
						"domain_prefix": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceLdsNetStorageGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsNetStorageGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id":            {Type: schema.TypeString, Computed: true},
			"cp_code":       {Type: schema.TypeString, Computed: true},
			"domain_prefix": {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataLdsNetStorageGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	data, err := api.ldsv3.ListNetStorageGroups()
	if err != nil {
		return diag.FromErr(err)
	}

	list := make([]map[string]interface{}, 0, len(*data))

	for _, item := range *data {
		object, err := flattenLdsNetStorageGroup(&item)
		if err != nil {
			return diag.FromErr(err)
		}
		list = append(list, object)
	}

	if err := d.Set("netstorage_groups", list); err != nil {
		return diag.FromErr(fmt.Errorf("%q", err.Error()))
	}

	jsonBody, err := json.Marshal(list)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSHAString(string(jsonBody)))

	return diags
}

func dataLdsNetStorageGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	fName := d.Get("name").(string)

	data, err := api.ldsv3.ListNetStorageGroups()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, item := range *data {
		if fName == item.ID {
			object, err := flattenLdsNetStorageGroup(&item)
			if err != nil {
				return diag.FromErr(err)
			}

			d.SetId(object["id"].(string))
			d.Set("cp_code", object["cp_code"].(string))
			d.Set("domain_prefix", object["domain_prefix"].(string))

			return diags
		}
	}

	return diags
}
