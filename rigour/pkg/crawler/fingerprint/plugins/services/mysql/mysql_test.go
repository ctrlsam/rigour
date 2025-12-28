package mysql

import (
	"testing"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/crawler/test"
	"github.com/ory/dockertest/v3"
)

func TestMySQL(t *testing.T) {
	testcases := []test.Testcase{
		{
			Description: "mysql",
			Port:        3306,
			Protocol:    plugins.TCP,
			Expected: func(res *plugins.Service) bool {
				return res != nil
			},
			RunConfig: dockertest.RunOptions{
				Repository: "mysql",
				Tag:        "5.7.39",
				Env: []string{
					"MYSQL_ROOT_PASSWORD=my-secret-pw",
				},
			},
		},
	}

	p := &MYSQLPlugin{}

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
