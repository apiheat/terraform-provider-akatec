package akatec

import (
	"context"
	"fmt"
	"strconv"

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
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"expired",
					"suspended",
				}, false),
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
					"delivery_threshold_id",
				},
			},
			"delivery_frequency_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_residual_data": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"delivery_frequency_id",
				},
			},
			"delivery_threshold_id": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"delivery_frequency_id",
				},
			},
			"delivery_threshold_name": {
				Type:     schema.TypeString,
				Computed: true,
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
		CreateContext:  resourceLdsConfigurationCreateCtx,
		ReadContext:    resourceLdsConfigurationReadCtx,
		UpdateContext:  resourceLdsConfigurationUpdateCtx,
		DeleteContext:  resourceLdsConfigurationDeleteCtx,
		StateUpgraders: []schema.StateUpgrader{},
		Exists:         resourceConfigurationExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: "",
		Timeouts:           &schema.ResourceTimeout{},
		Description:        "",
		UseJSONNumber:      false,
	}
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

func resourceLdsConfigurationCreateCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if d.Get("status").(string) == "expired" {
		return diag.FromErr(fmt.Errorf("During lds configuration creation 'status' can be only 'active' or 'suspended'"))
	}

	api := m.(*AkamaiServices)

	body, err := setBody(d)
	if err != nil {
		return diag.FromErr(err)

	}

	id, err := api.ldsv3.CreateLogConfiguration(
		body.LogSource.ID,
		body.LogSource.Type,
		body)
	if err != nil {
		return diag.FromErr(err)

	}

	d.SetId(id)

	// Now we need to handle status of log delivery
	if d.Get("status").(string) == "suspended" {
		err := api.ldsv3.SuspendLogConfiguration(id)
		if err != nil {
			return diag.FromErr(err)

		}
	}

	resourceLdsConfigurationReadCtx(ctx, d, m)

	return diags
}

func resourceLdsConfigurationUpdateCtx(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// if d.Get("status").(string) == "expired" {
	// 	return diag.FromErr(fmt.Errorf("During lds configuration update 'status' can be only 'active' or 'suspended'"))
	// }

	api := m.(*AkamaiServices)

	body, err := setBody(d)
	if err != nil {
		return diag.FromErr(err)

	}

	id := d.Id()

	_, err = api.ldsv3.UpdateLogConfiguration(id, body)
	if err != nil {
		return diag.FromErr(err)

	}

	// Now we need to handle status of log delivery
	if d.HasChange("status") {
		switch d.Get("status").(string) {
		case "active":
			err := api.ldsv3.ResumeLogConfiguration(id)
			if err != nil {
				return diag.FromErr(err)

			}
		case "suspended":
			err := api.ldsv3.SuspendLogConfiguration(id)
			if err != nil {
				return diag.FromErr(err)

			}
		case "expired":
			return diag.FromErr(fmt.Errorf("Expired status cannot be set by user"))
		}
	}

	d.SetId(id)

	resourceLdsConfigurationReadCtx(ctx, d, m)

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

	if err := d.Set("end_date", configuration.EndDate); err != nil {
		return diag.FromErr(err)
	}

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
	if err := d.Set("log_source_details", flattenLogSourceDetailsData(&configuration.LogSource)); err != nil {
		return diag.FromErr(err)
	}

	// Contact
	if err := d.Set("contact_details", flattenContactDetailsData(&configuration.ContactDetails)); err != nil {
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
	if err := d.Set("encoding_details", flattenEncodingDetailsData(&configuration.EncodingDetails)); err != nil {
		return diag.FromErr(err)
	}

	// Aggregation
	if err := d.Set("aggregation_type", configuration.AggregationDetails.Type); err != nil {
		return diag.FromErr(err)
	}
	switch configuration.AggregationDetails.Type {
	case "byLogArrival":
		if err := d.Set("delivery_frequency_id", configuration.AggregationDetails.DeliveryFrequency.ID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("delivery_frequency_name", configuration.AggregationDetails.DeliveryFrequency.Value); err != nil {
			return diag.FromErr(err)
		}
	case "byHitTime":
		if err := d.Set("deliver_residual_data", configuration.AggregationDetails.DeliverResidualData); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("delivery_threshold_id", configuration.AggregationDetails.DeliveryThreshold.ID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("delivery_threshold_name", configuration.AggregationDetails.DeliveryThreshold.Value); err != nil {
			return diag.FromErr(err)
		}
	}

	// Delivery
	if err := d.Set("delivery_type", configuration.DeliveryDetails.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("delivery_details", flattenDeliveryDetailsData(&configuration.DeliveryDetails)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(configuration.ID))

	return diags
}

func resourceConfigurationExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*AkamaiServices)

	id := d.Id()

	exists, err := c.ldsv3.GetLogConfiguration(id)

	return exists != nil, err
}
