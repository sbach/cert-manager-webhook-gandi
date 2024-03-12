package main

import (
	"os"
	"testing"

	"k8s.io/klog/v2/klogr"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
	runtimeLog "sigs.k8s.io/controller-runtime/pkg/log"
)

func init() {
	runtimeLog.SetLogger(klogr.New())
}

func TestAPIKey(t *testing.T) {
	fixture := acmetest.NewFixture(&GandiDNSProviderSolver{},
		acmetest.SetResolvedZone(os.Getenv("TEST_ZONE_NAME")),
		acmetest.SetManifestPath("./testdata/using-api-key"),
	)
	fixture.RunConformance(t)
}

func TestPersonalAccessToken(t *testing.T) {
	fixture := acmetest.NewFixture(&GandiDNSProviderSolver{},
		acmetest.SetResolvedZone(os.Getenv("TEST_ZONE_NAME")),
		acmetest.SetManifestPath("./testdata/using-personal-access-token"),
	)
	fixture.RunConformance(t)
}
