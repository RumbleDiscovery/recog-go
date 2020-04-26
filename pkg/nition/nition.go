package nition

//go:generate git submodule update
//go:generate go run vfsgen-recog/main.go

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	recog "github.com/hdm/recog-go"
)

// FingerprintSet is a collection of loaded Recog fingerprint databases
type FingerprintSet struct {
	Databases map[string]*recog.FingerprintDB
	Logger    *log.Logger
}

// NewFingerprintSet returns an allocated FingerprintSet structure
func NewFingerprintSet() *FingerprintSet {
	fs := &FingerprintSet{}
	fs.Databases = make(map[string]*recog.FingerprintDB)
	return fs
}

// MatchFirst matches data to a given fingerprint database
func (fs *FingerprintSet) MatchFirst(name string, data string) *recog.FingerprintMatch {
	nomatch := &recog.FingerprintMatch{Matched: false}
	fdb, ok := fs.Databases[name]
	if !ok {
		nomatch.Errors = append(nomatch.Errors, fmt.Errorf("database %s is missing", name))
		return nomatch
	}
	return fdb.MatchFirst(data)
}

// MatchAll matches data to a given fingerprint database
func (fs *FingerprintSet) MatchAll(name string, data string) []*recog.FingerprintMatch {
	nomatch := &recog.FingerprintMatch{Matched: false}
	fdb, ok := fs.Databases[name]
	if !ok {
		nomatch.Errors = append(nomatch.Errors, fmt.Errorf("database %s is missing", name))
		return []*recog.FingerprintMatch{nomatch}
	}
	return fdb.MatchAll(data)
}

// LoadFingerprints parses embedded Recog XML databases, returning a FingerprintSet
func (fs *FingerprintSet) LoadFingerprints() error {
	rootfs, err := Assets.Open("/")
	if err != nil {
		return err
	}
	defer rootfs.Close()

	files, err := rootfs.Readdir(65535)
	if err != nil {
		return err
	}

	for _, f := range files {

		if !strings.Contains(f.Name(), ".xml") {
			continue
		}

		fd, err := Assets.Open(f.Name())
		if err != nil {
			return err
		}

		xmlData, err := ioutil.ReadAll(fd)
		if err != nil {
			fd.Close()
			return err
		}
		fd.Close()

		fdb, err := recog.LoadFingerprintDB(f.Name(), xmlData)
		if err != nil {
			return err
		}

		fdb.Logger = fs.Logger

		// Create an alias for the file name
		fs.Databases[f.Name()] = &fdb

		// Create an alias for the "matches" attribute
		fs.Databases[fdb.Matches] = &fdb
	}

	return nil
}

// LoadFingerprintsDir parses Recog XML files from a local directory, returning a FingerprintSet
func (fs *FingerprintSet) LoadFingerprintsDir(dname string) error {
	files, err := ioutil.ReadDir(dname)
	if err != nil {
		return err
	}

	for _, f := range files {

		if !strings.Contains(f.Name(), ".xml") {
			continue
		}

		xmlData, err := ioutil.ReadFile(filepath.Join(dname, f.Name()))
		if err != nil {
			return err
		}

		fdb, err := recog.LoadFingerprintDB(f.Name(), xmlData)
		if err != nil {
			return err
		}

		fdb.Logger = fs.Logger

		// Create an alias for the file name
		fs.Databases[f.Name()] = &fdb

		// Create an alias for the "matches" attribute
		fs.Databases[fdb.Matches] = &fdb
	}

	return nil
}

// LoadFingerprints parses embedded Recog XML databases, returning a FingerprintSet
func LoadFingerprints() (*FingerprintSet, error) {
	res := NewFingerprintSet()
	return res, res.LoadFingerprints()
}

// LoadFingerprintsDir parses Recog XML files from a local directory, returning a FingerprintSet
func LoadFingerprintsDir(dname string) (*FingerprintSet, error) {
	res := NewFingerprintSet()
	return res, res.LoadFingerprintsDir(dname)
}

// MustLoadFingerprints loads the built-in fingerprints, panicing otherwise
func MustLoadFingerprints() *FingerprintSet {
	fset, err := LoadFingerprints()
	if err != nil {
		panic(err)
	}
	return fset
}
