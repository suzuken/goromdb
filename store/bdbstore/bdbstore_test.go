package bdbstore

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/goromdb/testutil"
)

var sampleDBFile = "../../data/store/sample-bdb.db"

func TestNew(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	filein := make(chan string)
	buf := new(bytes.Buffer)
	logger := log.New(buf, "", 0)
	_, err := New(filein, dir, logger)

	assert.Nil(t, err)
}

func TestLoad(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	filein := make(chan string)
	buf := new(bytes.Buffer)
	logger := log.New(buf, "", 0)
	s, _ := New(filein, dir, logger)

	type Case struct {
		input       string
		expectError bool
		subtest     string
	}
	cases := []Case{
		{
			sampleDBFile + ".hoge",
			true,
			"non-existing file fails",
		},
		{
			"../../data/store/sample-data.json",
			true,
			"non-bdb file fails",
		},
		{
			sampleDBFile,
			false,
			"existing bdb file succeeds",
		},
		{
			sampleDBFile,
			false,
			"another bdb file succeeds",
		},
	}

	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			err := s.Load(c.input)
			assert.Equal(t, c.expectError, err != nil)
		})
	}
}

func TestGet(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	filein := make(chan string)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	s, _ := New(filein, dir, logger)
	s.Load(sampleDBFile)

	type Case struct {
		input       string
		expectedVal []byte
		expectError bool
		subtest     string
	}
	cases := []Case{
		{
			"hoge",
			[]byte("hoge!"),
			false,
			"existing key returns expected val",
		},
		{
			"hogehoge",
			nil,
			true,
			"non-existing key returns error",
		},
	}

	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			val, err := s.Get([]byte(c.input))

			assert.Equal(t, c.expectError, err != nil)
			assert.Equal(t, c.expectedVal, val)
		})
	}
}

func TestStart(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	filein := make(chan string)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	s, _ := New(filein, dir, logger)
	done := s.Start()

	file := filepath.Join(dir, "dropin.db")
	for i := 0; i < 10; i++ {
		testutil.CopyFile(file, sampleDBFile)
		filein <- file
	}

	val, err := s.Get([]byte("hoge"))

	assert.Nil(t, err)
	assert.Equal(t, "hoge!", string(val))

	close(filein)
	<-done
}
