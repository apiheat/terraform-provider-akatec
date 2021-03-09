package akatec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceLdsEncodings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsEncodingsRead,
		Description: descriptions["encodings"],
		Schema: map[string]*schema.Schema{
			"log_source_type": {
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
			"delivery_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["delivery_type"],
				ValidateFunc: validation.StringInSlice([]string{
					"email",
					"ftp",
					"httpsns4",
				}, false),
			},
			"encodings": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: descriptions["encoding"],
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

func dataSourceLdsEncoding() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsEncodingRead,
		Description: descriptions["encoding"],
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_source_type": {
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
			"delivery_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["delivery_type"],
				ValidateFunc: validation.StringInSlice([]string{
					"email",
					"ftp",
					"httpsns4",
				}, false),
			},
			"id": {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataLdsEncodingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type
	var deliveryType string

	lgsType := d.Get("log_source_type").(string)
	dType, ok := d.GetOk("delivery_type")
	if ok {
		deliveryType = dType.(string)
	}

	data, err := api.ldsv3.ListLogEncodingsByType(lgsType, deliveryType)
	if err != nil {
		return diag.FromErr(err)
	}

	list := make([]map[string]interface{}, 0, len(*data))

	for _, item := range *data {
		list = append(list, flattenLdsParameter(&item))
	}

	if err := d.Set("encodings", list); err != nil {
		return diag.FromErr(fmt.Errorf("%q", err.Error()))
	}

	jsonBody, err := json.Marshal(list)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSHAString(string(jsonBody)))

	return diags
}

func dataLdsEncodingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type
	var deliveryType string

	lgsType := d.Get("log_source_type").(string)
	dType, ok := d.GetOk("delivery_type")
	if ok {
		deliveryType = dType.(string)
	}

	fName := d.Get("name").(string)

	data, err := api.ldsv3.ListLogEncodingsByType(lgsType, deliveryType)
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
