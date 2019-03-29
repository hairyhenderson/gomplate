//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/boltdb/bolt"
	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
)

type BoltDBDatasourcesSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&BoltDBDatasourcesSuite{})

func (s *BoltDBDatasourcesSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests")
	db, err := bolt.Open(s.tmpDir.Join("config.db"), 0600, nil)
	handle(c, err)
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
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
	result := icmd.RunCommand(GomplateBin,
		"-d", "config=boltdb://"+s.tmpDir.Join("config.db#Bucket1"),
		"-i", `{{(ds "config" "foo")}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config=boltdb://"+s.tmpDir.Join("config.db#Bucket1"),
		"-d", "config2=boltdb://"+s.tmpDir.Join("config.db#Bucket2"),
		"-i", `{{(ds "config" "foo")}}-{{(ds "config2" "foobar")}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar-baz"})
}
