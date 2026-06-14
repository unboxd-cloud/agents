package cloudstack

// Mapping describes how Apache CloudStack primitives map into Unboxd Platform
// product, tenant, metering, billing, and compliance records.
type Mapping struct {
	Anchor     string             `json:"anchor"`
	Provider   string             `json:"provider"`
	Primitives []PrimitiveMapping `json:"primitives"`
	Flows      []Flow             `json:"flows"`
}

type PrimitiveMapping struct {
	CloudStack string `json:"cloudstack"`
	Unboxd     string `json:"unboxd"`
	Purpose    string `json:"purpose"`
}

type Flow struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

func DefaultMapping() Mapping {
	return Mapping{
		Anchor:   "Apache CloudStack",
		Provider: "cloudstack",
		Primitives: []PrimitiveMapping{
			{CloudStack: "zone", Unboxd: "region", Purpose: "provider location and residency boundary"},
			{CloudStack: "pod", Unboxd: "availability boundary", Purpose: "physical or operational failure boundary"},
			{CloudStack: "cluster", Unboxd: "compute pool", Purpose: "capacity group"},
			{CloudStack: "host", Unboxd: "capacity resource", Purpose: "physical or virtual capacity provider"},
			{CloudStack: "domain", Unboxd: "organization boundary", Purpose: "tenant, reseller, or governance boundary"},
			{CloudStack: "account", Unboxd: "tenant account", Purpose: "billable and permissioned account"},
			{CloudStack: "project", Unboxd: "workspace", Purpose: "shared customer or internal workspace"},
			{CloudStack: "role", Unboxd: "permission boundary", Purpose: "access control mapping"},
			{CloudStack: "template", Unboxd: "image catalog item", Purpose: "approved base image"},
			{CloudStack: "service offering", Unboxd: "compute SKU", Purpose: "priced compute shape"},
			{CloudStack: "disk offering", Unboxd: "storage SKU", Purpose: "priced storage shape"},
			{CloudStack: "network offering", Unboxd: "network SKU", Purpose: "priced network policy"},
			{CloudStack: "event", Unboxd: "audit and metering source", Purpose: "operations and usage evidence"},
		},
		Flows: []Flow{
			{Name: "inventory", Steps: []string{"CloudStack inventory", "normalized records", "catalog", "dashboard"}},
			{Name: "usage", Steps: []string{"CloudStack usage/events", "metering", "rating", "billing"}},
			{Name: "compliance", Steps: []string{"CloudStack zones/domains", "residency policy", "evidence records", "reports"}},
		},
	}
}
