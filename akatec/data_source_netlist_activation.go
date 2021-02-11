package akatec

import (
	"context"
	"sort"

	svcNetList "github.com/apiheat/go-edgegrid/v6/service/netlistv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetlistActivation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetlistIPRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"acg": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contract_id": {
				Type:     schema.TypeString,
				Optional: true,
				RequiredWith: []string{
					"group_id",
				},
			},
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network": {
				Type:     schema.TypeString,
				Computed: true,
				//ValidateFunc: schema.SchemaValidateFunc(validation.StringInSlice([]string{"staging", "production"}, true)),
			},
			"activate": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					// ValidateDiagFunc: func(interface{}, cty.Path) diag.Diagnostics {
					// 	return nil
					// },
				},
			},
		},
	}
}

func dataSourceNetlistActivationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics // Warning or errors can be collected in a slice type

	netlistID := d.Get("id").(string)

	netlistOpts := svcNetList.ListNetworkListsOptionsv2{
		TypeOflist:      "IP",
		Extended:        true,
		IncludeElements: true,
		Search:          "",
	}

	netlist, err := api.netlistV2.GetNetworkList(netlistID, netlistOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", netlist.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("activate", false); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("network", "staging"); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", netlist.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("acg", netlist.AccessControlGroup); err != nil {
		return diag.FromErr(err)
	}

	sort.Strings(netlist.List)
	if err := d.Set("cidr_blocks", netlist.List); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(netlist.UniqueID)

	return diags
}
