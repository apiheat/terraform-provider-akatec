package akatec

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getSHAString(rdata string) string {
	h := sha1.New()
	h.Write([]byte(rdata))

	sha1hashtest := hex.EncodeToString(h.Sum(nil))
	return sha1hashtest
}

func flattenLogSourceDetailsData(l *ldsv3.OutputSourcesElement) map[string]interface{} {
	lgs := make(map[string]interface{})

	// Required fields
	if l.ID != "" {
		lgs["id"] = l.ID
	}
	if l.ID != "" {
		lgs["type"] = l.Type
	}
	if l.LogRetentionDays != 0 {
		lgs["log_retention"] = strconv.Itoa(l.LogRetentionDays)
	}

	// CP Code typed
	if l.CpCode != "" {
		lgs["cp_code"] = l.CpCode
	}
	if len(l.Products) != 0 {
		sortedProducts := l.Products
		sort.Strings(sortedProducts)
		lgs["products"] = strings.Join(sortedProducts, ",")
	}
	// EDNS
	if l.ZoneName != "" {
		lgs["zone_name"] = l.ZoneName
	}
	// AnswerX typed
	if l.Name != "" {
		lgs["name"] = l.Name
	}
	// GTM typed
	if l.PropertyName != "" {
		lgs["property_name"] = l.PropertyName
	}

	return lgs
}

func flattenLdsParameter(l *ldsv3.GenericConfigurationParameterElement) map[string]interface{} {
	lgs := make(map[string]interface{})

	lgs["id"] = l.ID
	lgs["name"] = l.Value

	return lgs
}

func flattenLdsNetStorageGroup(l *ldsv3.GenericConfigurationParameterElement) (map[string]interface{}, error) {
	lgs := make(map[string]interface{})

	lgs["id"] = l.ID
	lgs["name"] = l.ID

	nsURL, err := url.Parse(l.Value)
	if err != nil {
		return lgs, err
	}

	lgs["cp_code"] = path.Base(nsURL.Path)
	lgs["domain_prefix"] = strings.Split(l.Value, ".")[0]

	return lgs, nil
}

func flattenContactDetailsData(l *ldsv3.ConfigurationBodyContactDetails) map[string]interface{} {
	lgs := make(map[string]interface{})

	if l.Contact.ID != "" {
		lgs["id"] = l.Contact.ID
	}

	if len(l.MailAddresses) != 0 {
		sorted := l.MailAddresses
		sort.Strings(sorted)
		lgs["email_addresses"] = strings.Join(sorted, ",")
	}

	return lgs
}

func flattenEncodingDetailsData(l *ldsv3.ConfigurationBodyEncodingDetails) map[string]interface{} {
	lgs := make(map[string]interface{})

	if l.Encoding.ID != "" {
		lgs["id"] = l.Encoding.ID
	}

	if l.Encoding.ID == "4" {
		if l.EncodingKey != "" {
			lgs["encoding_key"] = l.EncodingKey
		}
	}

	return lgs
}

func flattenDeliveryDetailsData(l *ldsv3.ConfigurationBodyDeliveryDetails) map[string]interface{} {
	lgs := make(map[string]interface{})

	switch l.Type {
	case "httpsns4":
		if l.DomainPrefix != "" {
			lgs["domain_prefix"] = l.DomainPrefix
		}
		if l.Directory != "" {
			lgs["directory"] = l.Directory
		}
		idStr := strconv.Itoa(l.CpcodeID)
		if idStr != "" {
			lgs["cp_code"] = idStr
		}
	case "email":
		if l.EmailAddress != "" {
			lgs["email_address"] = l.EmailAddress
		}
	case "ftp":
		if l.Directory != "" {
			lgs["directory"] = l.Directory
		}
		if l.Login != "" {
			lgs["login"] = l.Login
		}
		if l.Machine != "" {
			lgs["machine"] = l.Machine
		}
		if l.Password != "" {
			lgs["password"] = l.Password
		}
	}

	return lgs
}

func setBody(d *schema.ResourceData) (body ldsv3.ConfigurationBody, errMsg error) {
	logSource := ldsv3.LogSourceBodyMember{
		ID:   d.Get("log_source_id").(string),
		Type: d.Get("log_source_type").(string),
	}

	contactObj := d.Get("contact_details").(map[string]interface{})
	contactDetails := ldsv3.ConfigurationBodyContactDetails{
		MailAddresses: strings.Split(contactObj["email_addresses"].(string), ","),
		Contact: ldsv3.GenericConfigurationParameterElement{
			ID: contactObj["id"].(string),
		},
	}

	logFormatDetails := ldsv3.ConfigurationBodyLogFormatDetails{
		LogIdentifier: d.Get("log_format_identifier").(string),
		LogFormat: ldsv3.GenericConfigurationParameterElement{
			ID: d.Get("log_format_id").(string),
		},
	}

	messageSize := ldsv3.GenericConfigurationParameterElement{
		ID: d.Get("message_size_id").(string),
	}

	aggrType := d.Get("aggregation_type").(string)
	aggregationDetails := ldsv3.ConfigurationBodyAggregationDetails{
		Type: aggrType,
	}
	switch aggrType {
	case "byLogArrival":
		freqID, getOk := d.GetOk("delivery_frequency_id")

		if !getOk {
			return body, fmt.Errorf("Missing required field `delivery_frequency_id` for aggregation type 'byLogArrival'")
		}

		aggregationDetails.DeliveryFrequency = &ldsv3.GenericConfigurationParameterElement{
			ID: freqID.(string),
		}
	case "byHitTime":
		dThr, getdT := d.GetOk("delivery_threshold_id")
		dRdt, getdR := d.GetOk("delivery_residual_data")

		if !getdT || !getdR {
			err := fmt.Errorf("Missing one or both required fields 'delivery_threshold_id' or 'delivery_residual_data' for aggregation type 'byHitTime'")
			return body, err
		}

		aggregationDetails.DeliveryThreshold = &ldsv3.GenericConfigurationParameterElement{
			ID: dThr.(string),
		}
		aggregationDetails.DeliverResidualData = dRdt.(bool)
	default:
		return body, fmt.Errorf("Unsupported aggregation type")
	}

	encodingObj := d.Get("encoding_details").(map[string]interface{})
	encodingDetails := ldsv3.ConfigurationBodyEncodingDetails{
		Encoding: ldsv3.GenericConfigurationParameterElement{
			ID: encodingObj["id"].(string),
		},
	}

	if encodingObj["id"].(string) == "4" {
		if key, getK := encodingObj["encoding_key"].(string); getK {
			encodingDetails.EncodingKey = key

			if key == "" {
				return body, fmt.Errorf("Encoding key cannot be empty")
			}

		}
	}

	dType := d.Get("delivery_type").(string)
	deliveryDetails := ldsv3.ConfigurationBodyDeliveryDetails{
		Type: dType,
	}

	deliveryObj := d.Get("delivery_details").(map[string]interface{})

	switch dType {
	case "email":
		deliveryDetails.EmailAddress = deliveryObj["email_address"].(string)
	case "ftp":
		deliveryDetails.Login = deliveryObj["login"].(string)
		deliveryDetails.Password = deliveryObj["password"].(string)
		deliveryDetails.Machine = deliveryObj["machine"].(string)
		deliveryDetails.Directory = deliveryObj["directory"].(string)
	case "httpsns4":
		delCpCodeInt, err := strconv.Atoi(deliveryObj["cp_code"].(string))
		if err != nil {
			return body, fmt.Errorf("Cannot convert delivery CP Code string value to integer")
		}
		deliveryDetails.CpcodeID = delCpCodeInt
		deliveryDetails.Directory = deliveryObj["directory"].(string)
		deliveryDetails.DomainPrefix = deliveryObj["domain_prefix"].(string)
	default:
		return body, fmt.Errorf("Unsupported delivery type")
	}

	body = ldsv3.ConfigurationBody{
		StartDate:          d.Get("start_date").(string),
		LogSource:          &logSource,
		ContactDetails:     contactDetails,
		LogFormatDetails:   logFormatDetails,
		MessageSize:        messageSize,
		AggregationDetails: aggregationDetails,
		EncodingDetails:    encodingDetails,
		DeliveryDetails:    deliveryDetails,
	}

	if d.Get("end_date").(string) != "" {
		body.EndDate = d.Get("end_date").(string)
	}

	return body, nil
}
