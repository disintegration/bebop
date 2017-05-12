package mysql

import "testing"

const (
	testAddress  = "127.0.0.1:3306"
	testUsername = "bebop_test_user"
	testPassword = "bebop_test_password"
	testDatabase = "bebop_test_database"
)

func getTestStore(t *testing.T) (*Store, func()) {
	s, err := Connect(testAddress, testUsername, testPassword, testDatabase)
	if err != nil {
		t.Fatalf(
			"failed to connect to the test mysql database: address=%q, username=%q, password=%q, database=%q: %s",
			testAddress, testUsername, testPassword, testDatabase, err,
		)
	}

	err = s.Reset()
	if err != nil {
		t.Fatalf("failed to reset the test mysql database: %s", err)
	}

	teardown := func() {
		err = s.Drop()
		if err != nil {
			t.Fatalf("failed to drop tables of the test mysql database: %s", err)
		}
	}

	return s, teardown
}

func TestPlaceholders(t *testing.T) {
	testTable := map[int]string{
		0:  "",
		1:  "?",
		2:  "?,?",
		5:  "?,?,?,?,?",
		-1: "",
	}
	for n, want := range testTable {
		got := placeholders(n)
		if got != want {
			t.Fatalf("got placeholders %q for %d, want %q", got, n, want)
		}
	}
}
