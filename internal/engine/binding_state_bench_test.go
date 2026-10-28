package engine

import "testing"

// Benchmarks intentionally exclude parsing and SSH transport. Compare these
// values with legacy RenderFrame and real terminal output before making
// performance claims about the complete application.
func BenchmarkBindingStateApply(b *testing.B) {
	state, err := NewBindingState(map[string]Value{"metrics.count": IntValue(0)})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := state.Apply(BindingUpdate{Key: "metrics.count", Value: IntValue(int64(i + 1))}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTypedFrameRender(b *testing.B) {
	view, err := Parse([]byte("<screen name=\"benchmark\"><text>Requests: {{metrics.count}}</text><text>Online: {{metrics.online}}</text></screen>"))
	if err != nil {
		b.Fatal(err)
	}
	values := map[string]Value{"metrics.count": IntValue(125), "metrics.online": BoolValue(true)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := view.RenderTypedFrame(values, 80); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLegacyFrameRender(b *testing.B) {
	view, err := Parse([]byte("<screen name=\"benchmark\"><text>Requests: {{metrics.count}}</text><text>Online: {{metrics.online}}</text></screen>"))
	if err != nil {
		b.Fatal(err)
	}
	values := map[string]string{"metrics.count": "125", "metrics.online": "true"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := view.RenderFrame(values, 80); err != nil {
			b.Fatal(err)
		}
	}
}
