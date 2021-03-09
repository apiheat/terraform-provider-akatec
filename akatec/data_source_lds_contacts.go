package akatec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLdsContacts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsContactsRead,
		Description: descriptions["contacts"],
		Schema: map[string]*schema.Schema{
			"contacts": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: descriptions["contact"],
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

func dataSourceLdsContact() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsContactRead,
		Description: descriptions["contact"],
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataLdsContactsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	data, err := api.ldsv3.ListContacts()
	if err != nil {
		return diag.FromErr(err)
	}

	list := make([]map[string]interface{}, 0, len(*data))

	for _, item := range *data {
		list = append(list, flattenLdsParameter(&item))
	}

	if err := d.Set("contacts", list); err != nil {
		return diag.FromErr(fmt.Errorf("%q", err.Error()))
	}

	jsonBody, err := json.Marshal(list)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSHAString(string(jsonBody)))

	return diags
}

func dataLdsContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	fName := d.Get("name").(string)

	data, err := api.ldsv3.ListContacts()
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
