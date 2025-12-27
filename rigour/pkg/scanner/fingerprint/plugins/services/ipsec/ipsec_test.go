package ipsec

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/scanner/test"
	"github.com/ory/dockertest/v3"
)

func TestIPSEC(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "ipsec",
			Port:        500,
			Protocol:    plugins.UDP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "hwdsl2/ipsec-vpn-server",
				Mounts: []string{
					"ikev2-vpn-data:/etc/ipsec.d",
					"/lib/modules:/lib/modules:ro",
				},
				Privileged: true,
			},
		},
	}

	var p *Plugin

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Description, func(t *testing.T) {
			t.Parallel()
			err := test.RunTest(t, tc, p)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
