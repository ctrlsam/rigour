package ntp

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/scanner/test"
	"github.com/ory/dockertest/v3"
)

func TestNTP(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "ntp",
			Port:        123,
			Protocol:    plugins.UDP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "cturra/ntp",
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
