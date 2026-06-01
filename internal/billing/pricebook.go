package billing

// SamplePriceBook returns a representative pay-as-you-go price book covering the
// meters used by the seeded catalog offerings, including AI meters. It
// demonstrates flat rates, free allowances, and graduated tiers in one book.
func SamplePriceBook() PriceBook {
	usd := "USD"
	return PriceBook{
		Version: "2026-06-01",
		Prices: map[string]MeterPrice{
			// Free monthly allowance, then graduated by usage.
			"compute.vcpu.hour": {Meter: "compute.vcpu.hour", Allowance: 100, Currency: usd,
				Tiers: []Tier{{UpTo: 1000, UnitPrice: 0.040}, {UpTo: 0, UnitPrice: 0.030}}},
			"compute.mem.gb.hour": {Meter: "compute.mem.gb.hour", Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.005}}},
			"storage.gb.month": {Meter: "storage.gb.month", Allowance: 50, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.020}}},
			"network.egress.gb": {Meter: "network.egress.gb", Allowance: 100, Currency: usd,
				Tiers: []Tier{{UpTo: 10000, UnitPrice: 0.090}, {UpTo: 0, UnitPrice: 0.050}}},
			"metrics.series.hour": {Meter: "metrics.series.hour", Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.0001}}},
			"messaging.msg.million": {Meter: "messaging.msg.million", Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.50}}},
			// AI-native, pay-as-you-go meters.
			"ai.gpu.hour": {Meter: "ai.gpu.hour", Currency: usd,
				Tiers: []Tier{{UpTo: 100, UnitPrice: 2.50}, {UpTo: 0, UnitPrice: 1.90}}},
			"ai.tokens.million": {Meter: "ai.tokens.million", Allowance: 1, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.60}}},
		},
	}
}
