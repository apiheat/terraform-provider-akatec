package akatec

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"

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
						"source_cpcode":    {Type: schema.TypeString, Computed: true, Optional: true},
						"source_products": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
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

	srcList := make([]map[string]interface{}, 0, len(*sources))

	for _, src := range *sources {
		if src.Type == "edns" {
			srcList = append(srcList, map[string]interface{}{
				"source_id":        src.ID,
				"source_type_name": src.Type,
				"source_cpcode":    "N/A",
				"source_products":  src.Products,
			})
		} else {
			srcList = append(srcList, map[string]interface{}{
				"source_id":        src.ID,
				"source_type_name": src.Type,
				"source_cpcode":    src.CpCode,
				"source_products":  src.Products,
			})
		}
	}

	if err := d.Set("sources", srcList); err != nil {
		return diag.FromErr(fmt.Errorf("%q", err.Error()))
	}

	jsonBody, err := json.Marshal(srcList)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSHAString(string(jsonBody)))

	return diags
}

func getSHAString(rdata string) string {
	h := sha1.New()
	h.Write([]byte(rdata))

	sha1hashtest := hex.EncodeToString(h.Sum(nil))
	return sha1hashtest
}
