package akatec

import (
	svcNetList "github.com/apiheat/go-edgegrid/v6/service/netlistv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetlistActivation() *schema.Resource {
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
		CreateContext:      resourceNetlistIPCreateCtx,
		ReadContext:        resourceNetlistIPReadCtx,
		UpdateContext:      resourceNetlistIPUpdateCtx,
		DeleteContext:      resourceNetlistIPDeleteCtx,
		StateUpgraders:     []schema.StateUpgrader{},
		Exists:             resourceNetlistIPExists,
		Importer:           &schema.ResourceImporter{},
		DeprecationMessage: "",
		Timeouts:           &schema.ResourceTimeout{},
		Description:        "",
		UseJSONNumber:      false,
	}
}

func resourceNetlistActivationExists(d *schema.ResourceData, m interface{}) (bool, error) {
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
