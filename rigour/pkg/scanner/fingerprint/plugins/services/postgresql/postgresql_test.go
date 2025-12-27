package postgres

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/scanner/test"
	"github.com/ory/dockertest/v3"
)

func TestPostgreSQL(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "postgresql",
			Port:        5432,
			Protocol:    plugins.TCP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "postgres",
				Env: []string{
					"POSTGRES_PASSWORD=secret",
					"POSTGRES_USER=user_name",
					"POSTGRES_DB=dbname",
					"listen_addresses = '*'",
				},
			},
		},
	}

	p := &POSTGRESPlugin{}

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
