package mqtt3

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/scanner/test"
	"github.com/ory/dockertest/v3"
)

func TestMqtt3(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "mqtt",
			Port:        1883,
			Protocol:    plugins.TCP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "efrecon/mosquitto",
			},
		},
	}

	p := &MQTT3Plugin{}

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
