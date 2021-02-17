package akatec

import (
	"github.com/apiheat/go-edgegrid/v6/service/ldsv3"
)

func flattenLdsParameter(l *ldsv3.GenericConfigurationParameterElement) map[string]interface{} {
	lgs := make(map[string]interface{})

	lgs["id"] = l.ID
	lgs["name"] = l.Value

	return lgs
}
