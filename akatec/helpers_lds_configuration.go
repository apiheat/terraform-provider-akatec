package akatec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	service "github.com/apiheat/go-edgegrid/v6/service/ldsv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
		// case "ftp":
		// 	if l.DeliveryDetails.DomainPrefix != "" {
		// 		lgs["domain_prefix"] = l.DeliveryDetails.DomainPrefix
		// 	}
		// 	if l.DeliveryDetails.Directory != "" {
		// 		lgs["directory"] = l.DeliveryDetails.Directory
		// 	}
		// 	idStr := strconv.Itoa(l.DeliveryDetails.CpcodeID)
		// 	if idStr != "" {
		// 		lgs["cp_code_id"] = idStr
		// 	}
	}

	return lgs
}

func setBody(d *schema.ResourceData) (body service.ConfigurationBody, errMsg error) {
	logSource := service.GenericBodyMember{
		ID:   d.Get("log_source_id").(string),
		Type: d.Get("log_source_type").(string),
	}

	contactObj := d.Get("contact_details").(map[string]interface{})
	contactDetails := service.ConfigurationBodyContactDetails{
		MailAddresses: strings.Split(contactObj["email_addresses"].(string), ","),
		Contact: service.GenericBodyMember{
			ID: contactObj["id"].(string),
		},
	}

	logFormatDetails := service.ConfigurationBodyLogFormatDetails{
		LogIdentifier: d.Get("log_format_identifier").(string),
		LogFormat: service.GenericBodyMember{
			ID: d.Get("log_format_id").(string),
		},
	}

	messageSize := service.GenericBodyMember{
		ID: d.Get("message_size_id").(string),
	}

	aggrType := d.Get("aggregation_type").(string)
	aggregationDetails := service.ConfigurationBodyAggregationDetails{
		Type: aggrType,
	}
	switch aggrType {
	case "byLogArrival":
		freqID, getOk := d.GetOk("delivery_frequency_id")

		if !getOk {
			return body, fmt.Errorf("Missing required field `delivery_frequency_id` for aggregation type 'byLogArrival'")
		}

		aggregationDetails.DeliveryFrequency = &service.GenericBodyMember{
			ID: freqID.(string),
		}
	case "byHitTime":
		dThr, getdT := d.GetOk("delivery_threshold")
		dRdt, getdR := d.GetOk("delivery_residual_data")

		if !getdT || !getdR {
			err := fmt.Errorf("Missing one or both required fields 'delivery_threshold' or 'delivery_residual_data' for aggregation type 'byHitTime'")
			return body, err
		}

		aggregationDetails.DeliveryThreshold = &service.GenericBodyMember{
			ID: dThr.(string),
		}
		aggregationDetails.DeliverResidualData = dRdt.(bool)
	default:
		return body, fmt.Errorf("Unsupported aggregation type")
	}

	encodingObj := d.Get("encoding_details").(map[string]interface{})
	encodingDetails := service.ConfigurationBodyEncodingDetails{
		Encoding: service.GenericBodyMember{
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
	deliveryDetails := service.ConfigurationBodyDeliveryDetails{
		Type: dType,
	}

	deliveryObj := d.Get("delivery_details").(map[string]interface{})

	switch dType {
	case "email":
		deliveryDetails.EmailAddress = ""
	case "ftp":
		deliveryDetails.Login = ""
		deliveryDetails.Password = ""
		deliveryDetails.Machine = ""
		deliveryDetails.Directory = ""
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

	body = service.ConfigurationBody{
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

	outputJSON(body)

	return body, nil
}

// ToDo: Remove debug function
func printJSON(str string) {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, []byte(str), "", "    ")
	if error != nil {
		log.Println("JSON parse error: ", error)
		return
	}
	log.Println(string(prettyJSON.Bytes()))
	return
}

// ToDo: Remove debug function
// OutputJSON displays output of query for alerts in JSON format
func outputJSON(input interface{}) {
	b, err := json.Marshal(input)
	if err != nil {
		log.Println(err)
		return
	}
	printJSON(string(b))
}
