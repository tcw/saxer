package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"gopkg.in/alecthomas/kingpin.v2"
	"bytes"
)

var (
	pathExp = kingpin.Arg("pathExp", "Sax Path Expression").Required().String()
	filename = kingpin.Arg("xml-file", "file").Required().String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	absFilename, err := abs(*filename)
	if err != nil {
		panic(err.Error())
	}
	SaxFile(absFilename)
}

func SaxFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	var count int64 = 0
	insideComment := false
	scanner.Split(elementSplit)
	for scanner.Scan() {
		text := scanner.Text()
		nospaceText := strings.TrimSpace(text)
		if insideComment {
			if strings.HasSuffix(nospaceText, "-->") {
				fmt.Println("Comment End: " + nospaceText)
				insideComment = false
			}
			continue
		}
		if nospaceText == "" {
		}else if strings.HasPrefix(nospaceText, "!--") && strings.HasSuffix(nospaceText, "-->") {
			// Begin commment  and end
			fmt.Println("Comment Start and end: " + nospaceText)
		}else if strings.HasPrefix(nospaceText, "!--") {
			// Begin commment
			insideComment = true
			fmt.Println("Comment Start: " + nospaceText)
		}else if strings.HasSuffix(nospaceText, "/>") {
			//is start and endelem
			fmt.Println("Start and end: " + nospaceText)
		}else if strings.HasPrefix(nospaceText, "/") {
			//is end elem
			fmt.Println("End: " + nospaceText)
		} else if strings.HasSuffix(nospaceText, ">") {
			//is startElem
			fmt.Println("Start: " + nospaceText)
		}else {
			//is startElem with content
			fmt.Println("Start with content: " + nospaceText)
		}
	}
	fmt.Printf("Read pages", count)
}



func elementSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "<"); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}


func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
