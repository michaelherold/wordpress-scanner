package shared

import (
	"archive/zip"
	"os"
)

type File struct {
	Path  string `json:"path"`
	Error error  `json:"error,omitempty"`
	Hash  uint32 `json:"hash"`
}

type Scan struct {
	Plugin  string `json:"plugin"`
	Version string `json:"version"`
	Files   []File `json:"files"`
}

func NewScan(plugin, version string) *Scan {
	return &Scan{plugin, version, []File{}}
}

func NewScanFromFile(plugin, version string, file *os.File) (*Scan, error) {
	scan := NewScan(plugin, version)

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	r, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, err
	}

	for _, f := range r.File {
		if f.Name[len(f.Name)-1] == '/' && f.UncompressedSize64 == 0 {
			continue
		}

		r, err := f.Open()
		if err != nil {
			scan.AddErrored(f.Name, err)
			continue
		}

		hash, err := GetHash(r)
		if err != nil {
			scan.AddErrored(f.Name, err)
			continue
		}
		r.Close()

		scan.AddHashed(f.Name, hash)
	}

	return scan, nil
}

func (s *Scan) AddHashed(path string, hash uint32) {
	s.Files = append(s.Files, File{path, nil, hash})
}

func (s *Scan) AddErrored(path string, err error) {
	s.Files = append(s.Files, File{path, err, 0})
}
