package rdp

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/crawler/test"
	"github.com/ory/dockertest/v3"
)

func TestRDP(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "rdp",
			Port:        3389,
			Protocol:    plugins.TCP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "scottyhardy/docker-remote-desktop",
			},
		},
	}

	p := &RDPPlugin{}

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
