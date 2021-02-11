package akatec

import (
	"context"
	"fmt"
	"sort"
	"strings"

	svcNetList "github.com/apiheat/go-edgegrid/v6/service/netlistv2"
	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetlistGeo() *schema.Resource {
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
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"network": {
				Type:     schema.TypeString,
				Default:  "staging",
				Optional: true,
				//ValidateFunc: schema.SchemaValidateDiagFunc(validation.StringInSlice([]string{"staging", "production"}),cty.Path{cty.GetAttrStep{Name: "foo"}},
			},
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"geo_codes": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validateGeoCode,
				},
			},
		},
		SchemaVersion:      0,
		CreateContext:      resourceNetlistGeoCreateCtx,
		ReadContext:        resourceNetlistGeoReadCtx,
		UpdateContext:      resourceNetlistGeoUpdateCtx,
		DeleteContext:      resourceNetlistGeoDeleteCtx,
		StateUpgraders:     []schema.StateUpgrader{},
		Exists:             resourceNetlistGeoExists,
		Importer:           &schema.ResourceImporter{},
		DeprecationMessage: "",
		Timeouts:           &schema.ResourceTimeout{},
		Description:        "",
		UseJSONNumber:      false,
	}
}

func resourceNetlistGeoDeleteCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	_, err := api.netlistV2.DeleteNetworkList(d.Id())
	if err != nil {
		return diag.FromErr(err)

	}

	return diags
}

func resourceNetlistGeoReadCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	netlistID := d.Id()

	netlistOpts := svcNetList.ListNetworkListsOptionsv2{
		TypeOflist:      "GEO",
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

	if err := d.Set("acg", netlist.AccessControlGroup); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("network", "staging"); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", netlist.Description); err != nil {
		return diag.FromErr(err)
	}

	sort.Strings(netlist.List)
	if err := d.Set("geo_codes", netlist.List); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(netlist.UniqueID)

	return diags
}

func resourceNetlistGeoCreateCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	netlistGroupId, netlistGroupIdExist := d.GetOk("group_id")
	if netlistGroupIdExist {
		netlistCreateOpts.GroupID = netlistGroupId.(int)
		netlistCreateOpts.ContractID = d.Get("contract_id").(string) // This attribute is required with group_id
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

	resourceNetlistGeoReadCtx(ctx, d, m)

	return diags
}

func resourceNetlistGeoUpdateCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	return resourceNetlistGeoReadCtx(ctx, d, m)
}

func resourceNetlistGeoExists(d *schema.ResourceData, m interface{}) (bool, error) {
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

func validateGeoCode(i interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	cty.IndexStringPath("geo_codes")
	attrVal := i.(string)

	for _, str := range []string{"AF", "AX", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW", "AU", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT", "BO", "BQ", "BA", "BW", "BV", "BR", "IO", "BN", "BG", "BF", "BI", "KH", "CM", "CA", "CV", "KY", "CF", "TD", "CL", "CN", "CX", "CC", "CO", "KM", "CG", "CD", "CK", "CR", "CI", "HR", "CU", "CW", "CY", "CZ", "DK", "DJ", "DM", "DO", "EC", "EG", "SV", "GQ", "ER", "EE", "ET", "FK", "FO", "FJ", "FI", "FR", "GF", "PF", "TF", "GA", "GM", "GE", "DE", "GH", "GI", "GR", "GL", "GD", "GP", "GU", "GT", "GG", "GN", "GW", "GY", "HT", "HM", "VA", "HN", "HK", "HU", "IS", "IN", "ID", "IR", "IQ", "IE", "IM", "IL", "IT", "JM", "JP", "JE", "JO", "KZ", "KE", "KI", "KP", "KR", "KW", "KG", "LA", "LV", "LB", "LS", "LR", "LY", "LI", "LT", "LU", "MO", "MK", "MG", "MW", "MY", "MV", "ML", "MT", "MH", "MQ", "MR", "MU", "YT", "MX", "FM", "MD", "MC", "MN", "ME", "MS", "MA", "MZ", "MM", "NA", "NR", "NP", "NL", "NC", "NZ", "NI", "NE", "NG", "NU", "NF", "MP", "NO", "OM", "PK", "PW", "PS", "PA", "PG", "PY", "PE", "PH", "PN", "PL", "PT", "PR", "QA", "RE", "RO", "RU", "RW", "BL", "SH", "KN", "LC", "MF", "PM", "VC", "WS", "SM", "ST", "SA", "SN", "RS", "SC", "SL", "SG", "SX", "SK", "SI", "SB", "SO", "ZA", "GS", "SS", "ES", "LK", "SD", "SR", "SJ", "SZ", "SE", "CH", "SY", "TW", "TJ", "TZ", "TH", "TL", "TG", "TK", "TO", "TT", "TN", "TR", "TM", "TC", "TV", "UG", "UA", "AE", "GB", "US", "UM", "UY", "UZ", "VU", "VE", "VN", "VG", "VI", "WF", "EH", "YE", "ZM", "ZW"} {
		if attrVal == str || (strings.ToLower(attrVal) == strings.ToLower(str)) {
			return diags
		}
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Geo country code incorrect",
			Detail:   fmt.Sprintf("Provided geo country code %s does not match any of the accepted values", attrVal),
		},
	}
}
