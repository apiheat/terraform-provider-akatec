package akatec

import (
	"context"
	"sort"

	svcNetList "github.com/apiheat/go-edgegrid/v6/service/netlistv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetlistIP() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: false,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "created by xakamai-tf",
			},
			"acg": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network": {
				Type:     schema.TypeString,
				Default:  "staging",
				Optional: true,
				//ValidateFunc: schema.SchemaValidateFunc(validation.StringInSlice([]string{"staging", "production"}, true)),
			},
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"cidr_blocks": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					// ValidateFunc: func(interface{}, string) ([]string, []error) {
					// 	return nil, nil
					// },
				},
			},
		},
		SchemaVersion:      0,
		CreateContext:      resourceNetlistCreateCtx,
		ReadContext:        resourceNetlistReadCtx,
		UpdateContext:      resourceNetlistUpdateCtx,
		DeleteContext:      resourceNetlistDeleteCtx,
		StateUpgraders:     []schema.StateUpgrader{},
		Exists:             resourceNetlistExists,
		Importer:           &schema.ResourceImporter{},
		DeprecationMessage: "",
		Timeouts:           &schema.ResourceTimeout{},
		Description:        "",
		UseJSONNumber:      false,
	}
}

func resourceNetlistDeleteCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	_, err := api.netlistV2.DeleteNetworkList(d.Id())
	if err != nil {
		return diag.FromErr(err)

	}

	return diags
}

func resourceNetlistReadCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	netlistID := d.Id()

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

func resourceNetlistCreateCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	cidrBlocks := []string{}

	for _, cidr := range d.Get("cidr_blocks").([]interface{}) {
		cidrBlocks = append(cidrBlocks, cidr.(string))
	}

	sort.Strings(cidrBlocks)

	netlistName := d.Get("name").(string)
	netlistType := "IP"
	netlistDesc := d.Get("description").(string)

	netlistCreateOpts := svcNetList.NetworkListsOptionsv2{
		Name:        netlistName,
		Type:        netlistType,
		Description: netlistDesc,
		List:        cidrBlocks,
	}

	newList, err := api.netlistV2.CreateNetworkList(netlistCreateOpts)
	if err != nil {

		netlistError := err.(*svcNetList.NetworkListErrorv2)

		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  netlistError.Title,
				Detail:   netlistError.Detail,
			},
		}
	}

	d.SetId(newList.UniqueID)

	resourceNetlistReadCtx(ctx, d, m)

	return diags
}

func resourceNetlistUpdateCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	netlistID := d.Id()

	netlistOpts := svcNetList.ListNetworkListsOptionsv2{
		TypeOflist:      "IP",
		Extended:        true,
		IncludeElements: true,
		Search:          "",
	}

	netlist, err := api.netlistV2.GetNetworkList(netlistID, netlistOpts)
	if err != nil {

		netlistError := err.(*svcNetList.NetworkListErrorv2)

		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  netlistError.Title,
				Detail:   netlistError.Detail,
			},
		}
	}

	if d.HasChange("description") {
		netlist.Description = d.Get("description").(string)
	}

	if d.HasChange("cidr_blocks") {
		cidrBlocks := []string{}
		for _, cidr := range d.Get("cidr_blocks").([]interface{}) {
			cidrBlocks = append(cidrBlocks, cidr.(string))
		}
		netlist.List = cidrBlocks
	}

	if d.HasChange("name") {
		netlist.Name = d.Get("name").(string)
	}

	_, err = api.netlistV2.ModifyNetworkList(*netlist)
	if err != nil {

		netlistError := err.(*svcNetList.NetworkListErrorv2)

		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  netlistError.Title,
				Detail:   netlistError.Detail,
			},
		}
	}

	return resourceNetlistReadCtx(ctx, d, m)
}

func resourceNetlistExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*AkamaiServices)

	netlistID := d.Id()

	netlistOpts := svcNetList.ListNetworkListsOptionsv2{
		TypeOflist:      "IP",
		Extended:        false,
		IncludeElements: false,
		Search:          "",
	}

	exists, err := c.netlistV2.GetNetworkList(netlistID, netlistOpts)

	return exists != nil, err
}
