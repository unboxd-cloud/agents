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
			// AWS-compatible service meters.
			"function.invocation.million": {Meter: "function.invocation.million", Allowance: 1, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.20}}},
			"function.gb.second": {Meter: "function.gb.second", Allowance: 400000, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.0000166667}}},
			"token.issued.million": {Meter: "token.issued.million", Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.10}}},
			"s3.request.million": {Meter: "s3.request.million", Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.40}}},
			"agent.run.hour": {Meter: "agent.run.hour", Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.12}}},
			"email.sent.thousand": {Meter: "email.sent.thousand", Allowance: 3, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.10}}},
			// AI-native, pay-as-you-go meters. CPU-based open-source LLMs are the
			// default and cheapest; GPU is optional and priced higher.
			"ai.cpu.hour": {Meter: "ai.cpu.hour", Allowance: 10, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.08}}},
			"ai.gpu.hour": {Meter: "ai.gpu.hour", Currency: usd,
				Tiers: []Tier{{UpTo: 100, UnitPrice: 2.50}, {UpTo: 0, UnitPrice: 1.90}}},
			"ai.tokens.million": {Meter: "ai.tokens.million", Allowance: 1, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.60}}},
			// Build stack (app builder) metered per build-minute.
			"build.minute": {Meter: "build.minute", Allowance: 500, Currency: usd,
				Tiers: []Tier{{UpTo: 0, UnitPrice: 0.008}}},
		},
	}
}
