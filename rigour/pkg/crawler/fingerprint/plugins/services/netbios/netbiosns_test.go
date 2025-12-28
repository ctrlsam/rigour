package netbios

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/crawler/test"
	"github.com/ory/dockertest/v3"
)

func TestNetBIOS(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "netbios-ns",
			Port:        137,
			Protocol:    plugins.UDP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "dperson/samba",
				Cmd:        []string{"-n"},
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
