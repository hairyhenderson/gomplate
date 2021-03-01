package integration

import (
	. "gopkg.in/check.v1"

	"go.etcd.io/bbolt"
	"gotest.tools/v3/fs"
)

type BoltDBDatasourcesSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&BoltDBDatasourcesSuite{})

func (s *BoltDBDatasourcesSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests")
	db, err := bbolt.Open(s.tmpDir.Join("config.db"), 0600, nil)
	handle(c, err)
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
	handle(c, err)
}

func (s *BoltDBDatasourcesSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *BoltDBDatasourcesSuite) TestBoltDBDatasource(c *C) {
	o, e, err := cmdTest(c, "-d", "config=boltdb://"+s.tmpDir.Join("config.db#Bucket1"),
		"-i", `{{(ds "config" "foo")}}`)
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdTest(c, "-d", "config=boltdb://"+s.tmpDir.Join("config.db#Bucket1"),
		"-d", "config2=boltdb://"+s.tmpDir.Join("config.db#Bucket2"),
		"-i", `{{(ds "config" "foo")}}-{{(ds "config2" "foobar")}}`)
	assertSuccess(c, o, e, err, "bar-baz")
}
