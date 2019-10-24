package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var showVerbose bool = false	// Show verbose output

func processFile(writer *bufio.Writer, path string) {
	// fmt.Printf("processFile - path: '%s'\n", path)
	tag := NewTag(path)
	ext := filepath.Ext(path)
	if ext == ".rake" || filepath.Base(path) == "Rakefile"  {
		ext = ".rb"
	}
	rset := Rules[ext]
	if rset != nil {
		if showVerbose {
			fmt.Println(path)
		}
		source, err := OpenLineSource(path)
		if err != nil {
			fmt.Println("Error opening file '" + path + "': " + err.Error())
			return
		}
		defer source.Close()
		for {
			line, err := source.ReadLine()
			if err != nil {
				break
			}
			rset.CheckLine(tag, line, source.Loc)
		}
		tag.WriteOn(writer)
	}
}

func walkDir(writer *bufio.Writer, path string, info os.FileInfo, err error) error {
	// fmt.Printf("walkDir -  path: '%s'\n", path)
	if info != nil && !info.IsDir() {
		if len(exclude) != 0 {
			// fmt.Printf("Exclude is: %s\n", exclude)
		}
		processFile(writer, path)
	}
	return nil
}

var version = "1.2.0"


type multiValueFlags []string

func (i *multiValueFlags) String() string {
	buf := ""
	for _, v := range *i {
		if buf != "" {
			buf += ", "
		}
		buf += v
	}
	return fmt.Sprintf("[%s]", buf)
}

func (i *multiValueFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	showVersion bool = false
	showHelp    bool = false
	exclude     multiValueFlags
	excludeDir  multiValueFlags
)

func main() {

	flag.BoolVar(&showVersion, "v",    false, "Display the version number")
	flag.BoolVar(&showVerbose, "V",    false, "Display files as processed")
	flag.BoolVar(&showHelp,    "h",    false, "Display help text")
	flag.BoolVar(&showHelp,    "help", false, "Display help text")
	flag.Var(&exclude,         "exclude", "Exclude files pattern")
	flag.Var(&excludeDir,      "exclude-dir", "Exclude dirs pattern")

	flag.Parse()

	fmt.Printf("excludeDir: %+v\n", excludeDir)
	// fmt.Printf("exclude: %+v\n", exclude)

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if showHelp {
		fmt.Println("Usage: gotags [options] [file...]")
		fmt.Println()
		fmt.Println("Options are:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	fo, _ := os.Create("TAGS")
	defer fo.Close()

	writer := bufio.NewWriter(fo)
	defer writer.Flush()

	walkFunc := func(path string, info os.FileInfo, err error) error {
		// fmt.Printf("walkFunc - path: '%s'\n", path)
		// fmt.Printf("walkDir - info: '%+v'\n", info)

		if info.IsDir() {

			if len(excludeDir) > 0 {

				// fmt.Printf("info.Name(): '%s'\n", info.Name())
				for _, excDir := range excludeDir {
					// fmt.Printf("excDir: '%s'\n", excDir)
					if excDir == info.Name() {
						fmt.Printf("Excluding dir '%s'\n", excDir)
						return filepath.SkipDir
					}
				}
			}

		}
		return walkDir(writer, path, info, err)
	}

	var err error = nil

	if len(flag.Args()) == 0 {
		err = filepath.Walk(".", walkFunc)
	} else {
		for _, arg := range flag.Args() {
			err = filepath.Walk(arg, walkFunc)
		}
	}
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(-1)
	}
}
