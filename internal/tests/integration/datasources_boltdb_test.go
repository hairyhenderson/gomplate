package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"gotest.tools/v3/fs"
)

func setupDatasourcesBoltDBTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests")
	t.Cleanup(tmpDir.Remove)

	db, err := bbolt.Open(tmpDir.Join("config.db"), 0o600, nil)
	require.NoError(t, err)
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b, err = tx.CreateBucket([]byte("Bucket1"))
		if err != nil {
			return err
		}
		// the first 8 bytes are ignored when read by libkv, so we prefix with gibberish
		err = b.Put([]byte("foo"), []byte("00000000bar"))
		if err != nil {
			return err
		}

		b, err = tx.CreateBucket([]byte("Bucket2"))
		if err != nil {
			return err
		}
		err = b.Put([]byte("foobar"), []byte("00000000baz"))
		return err
	})
	require.NoError(t, err)

	return tmpDir
}

func TestDatasources_BoltDB_Datasource(t *testing.T) {
	tmpDir := setupDatasourcesBoltDBTest(t)

	// ignore the stderr output, it'll contain boltdb deprecation warning
	o, _, err := cmd(t, "-d", "config=boltdb://"+tmpDir.Join("config.db#Bucket1"),
		"-i", `{{(ds "config" "foo")}}`).run()
	assertSuccess(t, o, "", err, "bar")

	o, _, err = cmd(t, "-d", "config=boltdb://"+tmpDir.Join("config.db#Bucket1"),
		"-d", "config2=boltdb://"+tmpDir.Join("config.db#Bucket2"),
		"-i", `{{(ds "config" "foo")}}-{{(ds "config2" "foobar")}}`).run()
	assertSuccess(t, o, "", err, "bar-baz")
}
