package akatec

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceLdsConfiguration() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_date": {
				Type:     schema.TypeString,
				Required: true,
			},
			"end_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aggregation_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"byLogArrival",
					"byHitTime",
				}, true),
			},
			"delivery_frequency_id": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"delivery_residual_data",
					"delivery_threshold",
				},
			},
			"delivery_frequency_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_residual_data": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"delivery_frequency_id",
				},
			},
			"delivery_threshold": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"delivery_frequency_id",
				},
			},
			"log_source_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_source_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"answerx",
					"cpcode-products",
					"edns",
					"gtm",
				}, false),
			},
			"log_source_details": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"contact_details": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"log_format_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_format_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_format_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message_size_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"message_size_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encoding_details": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"delivery_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"email",
					"ftp",
					"httpsns4",
				}, false),
			},
			"delivery_details": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		SchemaVersion:  0,
		CreateContext:  resourceLdsConfigurationReadCtx,
		ReadContext:    resourceLdsConfigurationReadCtx,
		UpdateContext:  resourceLdsConfigurationReadCtx,
		DeleteContext:  resourceLdsConfigurationReadCtx,
		StateUpgraders: []schema.StateUpgrader{},
		Exists:         resourceLdsConfigurationExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: "",
		Timeouts:           &schema.ResourceTimeout{},
		Description:        "",
		UseJSONNumber:      false,
	}
}

func flattenLogSourceDetailsData(l *ldsv3.ConfigurationsRespElem) map[string]interface{} {
	lgs := make(map[string]interface{})

	if l.LogSource.CpCode != "" {
		lgs["cp_code"] = l.LogSource.CpCode
	}
	if len(l.LogSource.Products) != 0 {
		sortedProducts := l.LogSource.Products
		sort.Strings(sortedProducts)
		lgs["products"] = strings.Join(sortedProducts, ",")
	}
	if l.LogSource.LogRetentionDays != 0 {
		lgs["log_retention"] = strconv.Itoa(l.LogSource.LogRetentionDays)
	}
	return lgs
}

func flattenContactDetailsData(l *ldsv3.ConfigurationsRespElem) map[string]interface{} {
	lgs := make(map[string]interface{})

	if l.ContactDetails.Contact.ID != "" {
		lgs["id"] = l.ContactDetails.Contact.ID
	}

	if len(l.ContactDetails.MailAddresses) != 0 {
		sorted := l.ContactDetails.MailAddresses
		sort.Strings(sorted)
		lgs["email_addresses"] = strings.Join(sorted, ",")
	}

	return lgs
}

func flattenEncodingDetailsData(l *ldsv3.ConfigurationsRespElem) map[string]interface{} {
	lgs := make(map[string]interface{})

	if l.EncodingDetails.Encoding.ID != "" {
		lgs["id"] = l.EncodingDetails.Encoding.ID
	}

	// ToDO: Uncomment encryption key later
	// if l.EncodingDetails.Encoding.ID == 4 {
	// 	if l.EncodingDetails.Encoding.EncodingKey != "" {
	// 		lgs["encoding_key"] = l.EncodingDetails.Encoding.EncodingKey
	// 	}
	// }

	return lgs
}

func flattenDeliveryDetailsData(l *ldsv3.ConfigurationsRespElem) map[string]interface{} {
	lgs := make(map[string]interface{})

	switch l.DeliveryDetails.Type {
	case "httpsns4":
		if l.DeliveryDetails.DomainPrefix != "" {
			lgs["domain_prefix"] = l.DeliveryDetails.DomainPrefix
		}
		if l.DeliveryDetails.Directory != "" {
			lgs["directory"] = l.DeliveryDetails.Directory
		}
		idStr := strconv.Itoa(l.DeliveryDetails.CpcodeID)
		if idStr != "" {
			lgs["cp_code"] = idStr
		}
	// case "email":
	// 	if l.DeliveryDetails.EmailAddress != "" {
	// 		lgs["email_address"] = l.DeliveryDetails.EmailAddress
	// 	}
	case "ftp":
		if l.DeliveryDetails.DomainPrefix != "" {
			lgs["domain_prefix"] = l.DeliveryDetails.DomainPrefix
		}
		if l.DeliveryDetails.Directory != "" {
			lgs["directory"] = l.DeliveryDetails.Directory
		}
		idStr := strconv.Itoa(l.DeliveryDetails.CpcodeID)
		if idStr != "" {
			lgs["cp_code_id"] = idStr
		}
	}

	return lgs
}

func resourceConfigurationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	id := d.Id()

	d.SetId(id)
	return []*schema.ResourceData{d}, nil
}

func resourceLdsConfigurationDeleteCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	err := api.ldsv3.RemoveLogConfiguration(d.Id())
	if err != nil {
		return diag.FromErr(err)

	}

	return diags
}

func resourceLdsConfigurationReadCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*AkamaiServices)

	var diags diag.Diagnostics

	id := d.Id()

	configuration, err := api.ldsv3.GetLogConfiguration(id)
	if err != nil {
		return diag.FromErr(err)

	}

	if err := d.Set("start_date", configuration.StartDate); err != nil {
		return diag.FromErr(err)
	}

	// if err := d.Set("end_date", configuration.EndDate); err != nil {
	// 	return diag.FromErr(err)
	// }

	if err := d.Set("status", configuration.Status); err != nil {
		return diag.FromErr(err)
	}

	// Log Source
	if err := d.Set("log_source_id", configuration.LogSource.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("log_source_type", configuration.LogSource.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", configuration.LogSource.CpCode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("log_source_details", flattenLogSourceDetailsData(configuration)); err != nil {
		return diag.FromErr(err)
	}

	// Contact
	if err := d.Set("contact_details", flattenContactDetailsData(configuration)); err != nil {
		return diag.FromErr(err)
	}

	// Log Format
	if err := d.Set("log_format_identifier", configuration.LogFormatDetails.LogIdentifier); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("log_format_id", configuration.LogFormatDetails.LogFormat.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("log_format_name", configuration.LogFormatDetails.LogFormat.Value); err != nil {
		return diag.FromErr(err)
	}

	// Message size
	if err := d.Set("message_size_id", configuration.MessageSize.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message_size_name", configuration.MessageSize.Value); err != nil {
		return diag.FromErr(err)
	}

	// Encodings
	if err := d.Set("encoding_details", flattenEncodingDetailsData(configuration)); err != nil {
		return diag.FromErr(err)
	}

	// Aggregation
	if err := d.Set("aggregation_type", configuration.AggregationDetails.Type); err != nil {
		return diag.FromErr(err)
	}
	switch configuration.AggregationDetails.Type {
	case "byLogArrival":
		if err := d.Set("delivery_frequency_details", configuration.AggregationDetails.DeliveryFrequency.ID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("delivery_frequency_name", configuration.AggregationDetails.DeliveryFrequency.Value); err != nil {
			return diag.FromErr(err)
		}
		// case "byHitTime":
		// 	lgs["deliver_residual_data"] = l.AggregationDetails.DeliverResidualData
		// 	if l.AggregationDetails.DeliveryThreshold.ID != "" {
		// 		lgs["id"] = l.AggregationDetails.DeliveryThreshold.ID
		// 	}
		// 	if l.AggregationDetails.DeliveryThreshold.Value != "" {
		// 		lgs["name"] = l.AggregationDetails.DeliveryThreshold.Value
		// 	}
	}

	// Delivery
	if err := d.Set("delivery_type", configuration.DeliveryDetails.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("delivery_details", flattenDeliveryDetailsData(configuration)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(configuration.ID))

	return diags
}

func resourceLdsConfigurationExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*AkamaiServices)

	id := d.Id()

	exists, err := c.ldsv3.GetLogConfiguration(id)

	return exists != nil, err
}
