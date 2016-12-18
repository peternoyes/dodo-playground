package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func Compile(source []byte, language string) ([]byte, error) {
	var libdir string
	var fileName string

	switch language {
	case "c":
		libdir = "lib"
		fileName = "main.c"
		break
	case "assembly":
		libdir = "lib-assembly"
		fileName = "main.s65"
		break
	}

	dir, err := ioutil.TempDir("", "build")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(dir)

	err = copyDir(libdir, dir)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(dir+"/"+fileName, source, 0644)
	if err != nil {
		return nil, err
	}

	errOutput := &bytes.Buffer{}

	cmd := exec.Command("make", "-C", dir)
	cmd.Stderr = errOutput

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = cmd.Run()
	if err != nil {
		errStr := string(errOutput.Bytes())
		newErr := errors.New(errStr)

		return nil, newErr
	}

	output, err := ioutil.ReadFile(dir + "/fram.bin")
	if err != nil {
		return nil, err
	}

	return output, nil
}

func copyFile(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err == nil {
		sourceInfo, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, sourceInfo.Mode())
		}
	}

	return err
}

func copyDir(source, dest string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceInfo.Mode())
	if err != nil {
		return err
	}

	directory, err := os.Open(source)
	if err != nil {
		return err
	}

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefile := source + "/" + obj.Name()
		destfile := dest + "/" + obj.Name()

		if obj.IsDir() {
			err = copyDir(sourcefile, destfile)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(sourcefile, destfile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
