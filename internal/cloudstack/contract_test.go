package cloudstack

import "testing"

func TestDeployVMRequest_Validate(t *testing.T) {
	valid := DeployVMRequest{
		Account: "t1", Name: "web-1", ZoneID: "zone-1",
		TemplateID: "tmpl-nginx", ServiceOfferingID: "so-small",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}

	missing := map[string]DeployVMRequest{
		"account":  {Name: "n", ZoneID: "z", TemplateID: "t", ServiceOfferingID: "s"},
		"name":     {Account: "a", ZoneID: "z", TemplateID: "t", ServiceOfferingID: "s"},
		"zoneid":   {Account: "a", Name: "n", TemplateID: "t", ServiceOfferingID: "s"},
		"template": {Account: "a", Name: "n", ZoneID: "z", ServiceOfferingID: "s"},
		"offering": {Account: "a", Name: "n", ZoneID: "z", TemplateID: "t"},
	}
	for field, req := range missing {
		if err := req.Validate(); err == nil {
			t.Errorf("expected validation error when %s is missing", field)
		}
	}
}
