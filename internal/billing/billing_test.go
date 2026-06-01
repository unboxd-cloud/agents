package billing

import (
	"testing"
	"time"

	"github.com/unboxd-cloud/platform/internal/metering"
)

func ev(tenant, meter string, qty float64) metering.UsageEvent {
	return metering.UsageEvent{TenantID: tenant, Meter: meter, Quantity: qty, At: time.Now()}
}

func TestRate_AllowanceAndGraduatedTiers(t *testing.T) {
	pb := SamplePriceBook()
	// 350 vCPU-hours: 100 free, 250 billable.
	// Tier1 upTo 1000 @0.04 covers all 250 -> 250*0.04 = 10.00
	events := []metering.UsageEvent{ev("t1", "compute.vcpu.hour", 350)}
	inv := Rate(pb, "t1", time.Time{}, time.Time{}, events)

	if len(inv.Lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(inv.Lines))
	}
	line := inv.Lines[0]
	if line.Billable != 250 {
		t.Errorf("billable: want 250, got %v", line.Billable)
	}
	if line.Amount != 10.00 {
		t.Errorf("amount: want 10.00, got %v", line.Amount)
	}
	if inv.Total != 10.00 {
		t.Errorf("total: want 10.00, got %v", inv.Total)
	}
}

func TestRate_CrossesTierBoundary(t *testing.T) {
	pb := SamplePriceBook()
	// 1300 vCPU-hours: 100 free -> 1200 billable.
	// 0..1000 @0.04 = 40.00 ; 1000..1200 @0.03 = 6.00 ; total 46.00
	events := []metering.UsageEvent{ev("t1", "compute.vcpu.hour", 1300)}
	inv := Rate(pb, "t1", time.Time{}, time.Time{}, events)
	if inv.Total != 46.00 {
		t.Errorf("total: want 46.00, got %v", inv.Total)
	}
}

func TestRate_FlatRate(t *testing.T) {
	pb := SamplePriceBook()
	// mem: no allowance, single unbounded tier @0.005. 2000 -> 10.00
	events := []metering.UsageEvent{ev("t1", "compute.mem.gb.hour", 2000)}
	inv := Rate(pb, "t1", time.Time{}, time.Time{}, events)
	if inv.Total != 10.00 {
		t.Errorf("total: want 10.00, got %v", inv.Total)
	}
}

func TestRate_AllowanceCoversAll(t *testing.T) {
	pb := SamplePriceBook()
	// storage allowance is 50; usage 40 -> 0 billable, 0 amount.
	events := []metering.UsageEvent{ev("t1", "storage.gb.month", 40)}
	inv := Rate(pb, "t1", time.Time{}, time.Time{}, events)
	if inv.Total != 0 {
		t.Errorf("total: want 0, got %v", inv.Total)
	}
}

func TestRate_TenantIsolation(t *testing.T) {
	pb := SamplePriceBook()
	events := []metering.UsageEvent{
		ev("t1", "compute.mem.gb.hour", 2000), // 10.00 for t1
		ev("t2", "compute.mem.gb.hour", 9999), // belongs to t2, must be ignored
	}
	inv := Rate(pb, "t1", time.Time{}, time.Time{}, events)
	if inv.Total != 10.00 {
		t.Errorf("tenant isolation broken: want 10.00, got %v", inv.Total)
	}
}

func TestRate_AggregatesSameMeter(t *testing.T) {
	pb := SamplePriceBook()
	events := []metering.UsageEvent{
		ev("t1", "compute.mem.gb.hour", 1000),
		ev("t1", "compute.mem.gb.hour", 1000),
	}
	inv := Rate(pb, "t1", time.Time{}, time.Time{}, events)
	if inv.Lines[0].Quantity != 2000 {
		t.Errorf("aggregation: want 2000, got %v", inv.Lines[0].Quantity)
	}
}

func TestSettle_ResellerMarkup(t *testing.T) {
	inv := Invoice{TenantID: "t1", Total: 100, Currency: "USD"}
	s := Settle(inv, Partner{ID: "p1", Mode: ModeReseller, Rate: 0.15})
	if s.GrossToCustomer != 115 {
		t.Errorf("gross: want 115, got %v", s.GrossToCustomer)
	}
	if s.NetToPlatform != 100 {
		t.Errorf("net: want 100, got %v", s.NetToPlatform)
	}
}

func TestSettle_MarketplaceCommission(t *testing.T) {
	inv := Invoice{TenantID: "t1", Total: 100, Currency: "USD"}
	s := Settle(inv, Partner{ID: "m1", Mode: ModeMarketplace, Rate: 0.10})
	if s.GrossToCustomer != 100 {
		t.Errorf("gross: want 100, got %v", s.GrossToCustomer)
	}
	if s.NetToPlatform != 90 {
		t.Errorf("net: want 90, got %v", s.NetToPlatform)
	}
}

func TestSettle_DirectPassthrough(t *testing.T) {
	inv := Invoice{TenantID: "t1", Total: 100, Currency: "USD"}
	s := Settle(inv, Partner{ID: "", Mode: ModeDirect})
	if s.GrossToCustomer != 100 || s.NetToPlatform != 100 {
		t.Errorf("direct passthrough broken: %+v", s)
	}
}
