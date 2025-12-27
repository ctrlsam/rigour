package http

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/scanner/test"
	"github.com/ory/dockertest/v3"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

func TestHTTPS(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "https",
			Port:        8443,
			Protocol:    plugins.TCPTLS,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "mendhak/http-https-echo",
				Tag:        "24",
			},
		},
	}

	p := HTTPSPlugin{}
	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		panic("unable to initialize wappalyzer library")
	}
	p.analyzer = wappalyzerClient

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Description, func(t *testing.T) {
			t.Parallel()
			err := test.RunTest(t, tc, &p)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
