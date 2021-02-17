package akatec

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceLdsSources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataLdsSourcesRead,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"answerx",
					"cpcode-products",
					"edns",
					"gtm",
				}, false),
			},
			"sources": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of logdelivery service sources",
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

func dataLdsSourcesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type
	var sources ldsv3.OutputSources

	lsType, exists := d.GetOk("type")

	if exists {
		data, err := api.ldsv3.ListSourcesByType(lsType.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		sources = *data
	} else {
		data, err := api.ldsv3.ListSources()
		if err != nil {
			return diag.FromErr(err)
		}
		sources = *data
	}

	srcList := make([]map[string]interface{}, 0, len(sources))

	for _, src := range sources {
		srcList = append(srcList, flattenLogSourceDetailsData(&src))
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
