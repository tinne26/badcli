package badcli

import "os"
import "errors"
import "strings"
import "path/filepath"

type FilePath struct {
	value string
	allowedExtensions []string
}

func NewFilePath(path string, allowedExtensions ...string) *FilePath {
	return &FilePath{
		value: path,
		allowedExtensions: allowedExtensions,
	}
}

func (self *FilePath) Value() string {
	return self.value
}

// Returns the file path as "directory/name.ext". This is usually short
// enough to be nice to print casually to console, and has a bit more
// context than the file name alone.
func (self *FilePath) Reference() string {
	dir := filepath.Base(filepath.Dir(self.value))
	return dir + string(os.PathSeparator) + filepath.Base(self.value)
}

func (self *FilePath) ParseFromArg(arg string) error {
	// empty value case
	if arg == "" { return ErrMissingValue }

	// weird suffix cases
	if hasAnySuffix(arg, ".", string(os.PathSeparator), string(os.PathListSeparator)) {
		return errors.New("given value doesn't look like a file path")
	}

	// check extension if we have an explicit list
	if len(self.allowedExtensions) > 0 {
		found := false
		for _, ext := range self.allowedExtensions {
			if strings.HasSuffix(arg, "." + ext) {
				found = true
				break
			}
		}
		if !found {
			if len(self.allowedExtensions) == 1 {
				// simple error message
				return errors.New("file path must end with '" + self.allowedExtensions[0] + "'")
			} else {
				// create nice error message
				var sep string = ", "
				var extInfos strings.Builder
				extInfos.WriteString("file path must end with ")
				for i, ext := range self.allowedExtensions {
					if i > 0 {
						last := (i == len(self.allowedExtensions) - 1)
						if last { sep = " or " }
						extInfos.WriteString(sep)
					}
					extInfos.WriteString(ext)
				}
				return errors.New(extInfos.String())
			}
		}
	}

	// try to obtain abs path or clean
	fullPath, err := filepath.Abs(arg)
	if err != nil { // fallback
		fullPath = filepath.Clean(arg)
	}

	// assign value and return
	self.value = fullPath
	return nil
}
