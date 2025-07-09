package gopopulate

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const validChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMONPQRSTUVWXYZ-_"

var fileTypes = make([]string, 5)
var src rand.Source

func init() {
	src = rand.NewSource(time.Now().UTC().UnixNano())
}

func PopulateDir(baseDir, types string, depth int, maxFiles int) error {
	err := parseTypes(types)
	if err != nil {
		return err
	}
	err = popDirHelper(baseDir, depth, maxFiles)
	if err != nil {
		return err
	}
	return nil
}

func parseTypes(types string) error {
	// make sure the slice is empty
	fileTypes = fileTypes[:0]
	for c := range types {
		switch types[c] {
		case 'd':
			if !contains(fileTypes, "dir") {
				fileTypes = append(fileTypes, "dir")
			}
		case 'r':
			if !contains(fileTypes, "regFile") {
				fileTypes = append(fileTypes, "regFile")
			}
		default:
			return errors.New("Unexpected type char: " + string(types[c]))
		}
	}
	return nil
}

func contains(arr []string, val string) bool {
	for _, a := range arr {
		if strings.Compare(a, val) == 0 {
			return true
		}
	}
	return false
}

func popDirHelper(baseDir string, depth, maxFiles int) error {
	for numFiles := 0; numFiles < maxFiles; numFiles++ {
		fileType := pickType()
		newFile, err := genFile(baseDir, fileType)
		if err != nil {
			return err
		}
		if strings.Compare(fileType, "dir") == 0 && depth > 0 {
			return popDirHelper(genFilePath(baseDir, newFile), depth-1, maxFiles)
		}
	}
	return nil
}

func pickType() string {
	return fileTypes[src.Int63()%(int64)(len(fileTypes))]
}

func genFile(baseDir, fileType string) (string, error) {
	var name string
	if strings.Compare(fileType, "dir") == 0 {
		name = genRandomName()
		if err := os.Mkdir(genFilePath(baseDir, name), 0777); err != nil {
			return "", err
		}
	} else if strings.Compare(fileType, "regFile") == 0 {
		name := genRandomName()
		path := genFilePath(baseDir, name)
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			file, err := os.Create(path)
			if err != nil {
				return "", err
			}
			defer file.Close()
			if err = writeDataToFile(file); err != nil {
				return "", err
			}
		}
	}
	return name, nil
}

func genFilePath(baseDir, newFile string) string {
	var buffer bytes.Buffer
	buffer.WriteString(baseDir)
	buffer.WriteString("/")
	buffer.WriteString(newFile)
	return buffer.String()
}

func genRandomName() string {
	var length int64
	for length == 0 {
		length = rand.Int63() % 8
	}
	return genRandomString(length)
}

func writeDataToFile(f *os.File) error {
	_, err := f.WriteString(genRandomDataForFile())
	if err != nil {
		return err
	}
	f.Sync()
	return nil
}

func genRandomDataForFile() string {
	lines := rand.Int63() % 64

	var buffer bytes.Buffer
	var i int64
	for i = 0; i < lines; i++ {
		length := rand.Int63() % 80
		buffer.WriteString(genRandomString(length))
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func genRandomString(length int64) string {
	name := make([]byte, length)

	for i := range name {
		name[i] = validChars[src.Int63()%int64(len(validChars))]
	}
	return string(name)
}

// Tar a directory and its content
func Tar(tarDir, tarPath string) error {
	file, err := os.Create(tarPath)
	if err != nil {
		return err
	}

	tw := tar.NewWriter(file)
	defer tw.Flush()
	defer tw.Close()
	paths, err := generateFilePaths(tarDir)
	if err != nil {
		return err
	}

	for _, file := range paths {
		if err := addToTar(tw, tarDir, file); err != nil {
			return err
		}
	}
	return nil
}

func generateFilePaths(basePath string) ([]string, error) {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, len(files))
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}

func addToTar(tw *tar.Writer, basePath, path string) error {
	file, err := os.Open(filepath.Join(basePath, path))
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return nil
	}

	header := new(tar.Header)
	header.Name = path
	header.Size = stat.Size()
	header.Mode = int64(stat.Mode())
	header.ModTime = stat.ModTime()

	if err = tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err = io.Copy(tw, file); err != nil {
		return err
	}
	return tw.Flush()
}
