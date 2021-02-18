package akatec

import (
	"net/url"
	"path"
	"strings"

	"github.com/apiheat/go-edgegrid/v6/service/ldsv3"
)

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
