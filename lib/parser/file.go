package parser

import (
	"errors"
	"fmt"
	"regexp"
)

type Files struct {
	Source      []string // absolute file paths
	Destination string   // absolute dir path
}

func FileParser(source string) (*Files, error) {
	files := regexp.MustCompile("\\s+").Split(source, -1)

	if len(files) < 2 {
		return nil, errors.New(fmt.Sprintf("COPY/DOWNLOAD requires a minimum of two parameters, now received %s", source))
	}

	lastElementIndex := len(files)

	sourceFiles := files[:lastElementIndex-1]
	destinationDir := files[lastElementIndex-1]

	return &Files{
		Source:      sourceFiles,
		Destination: destinationDir,
	}, nil
}
