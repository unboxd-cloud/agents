package plugin

import "testing"

func TestRegisterNewList(t *testing.T) {
	Register(Extension{
		Kind: KindProtocol,
		Name: "test-proto",
		Factory: func(cfg map[string]string) (any, error) {
			return cfg["v"], nil
		},
	})
	got, err := New(KindProtocol, "test-proto", map[string]string{"v": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if got.(string) != "hi" {
		t.Errorf("factory config not passed: %v", got)
	}
	found := false
	for _, n := range List(KindProtocol) {
		if n == "test-proto" {
			found = true
		}
	}
	if !found {
		t.Error("registered extension not listed")
	}
}

func TestNewUnknown(t *testing.T) {
	if _, err := New(KindProvider, "nope", nil); err == nil {
		t.Error("expected error for unknown extension")
	}
}
