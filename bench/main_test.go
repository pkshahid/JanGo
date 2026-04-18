package bench

import (
	"testing"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// Enable goleak globally for tests in this package.
	// We ignore monitoring background routine which is by-design.
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/pkshahid/JanGo/monitoring.updateRuntimeMetrics"))
}
