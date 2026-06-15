package adl

import (
	"os"
	"path/filepath"
	"testing"
)

func benchSource(b *testing.B) string {
	b.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sample.agent"))
	if err != nil {
		b.Fatal(err)
	}
	return string(data)
}

func BenchmarkParse(b *testing.B) {
	src := benchSource(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(src)
	}
}

func BenchmarkCompile(b *testing.B) {
	src := benchSource(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Compile(src)
	}
}

func BenchmarkLoadAgent(b *testing.B) {
	src := benchSource(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAgent(Compile(src).Model)
	}
}
