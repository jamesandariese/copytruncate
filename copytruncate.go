package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

var format = flag.String("format", "2006-02-01", "Time format")

func removeUnderlyingFile(name string, file *os.File) {
	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't stat", name, err)
		return
	}
	nfi, err := os.Stat(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't delete", name, err)
		return
	}
	if os.SameFile(fi, nfi) {
		os.Remove(name)
	}
}

func copyTruncate(nameIn, nameOut string) error {
	// copy with gzip and truncate a file.
	// filename may be relative or absolute.
	// this will cd to the directory with the filec.
	out, err := os.Create(nameOut)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	in, err := os.OpenFile(nameIn, os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer out.Close()
	defer in.Close()
	defer func() {
		// if the function fails, let's delete the file we were creating to save space.  OKAY!
		if r := recover(); r != nil {
			removeUnderlyingFile(nameOut, out)
		}
	}()

	gzout := gzip.NewWriter(out)
	defer gzout.Close()

	// Don't truncate until we've got 10 empty reads in a row
	for i := 0; i < 10; i++ {
		written, err := io.Copy(gzout, in)
		if written > 0 {
			i = 0
			fmt.Print(".")
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			panic(err)
		}
	}
	err = in.Truncate(0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't truncate file", err)
		panic(err)
	}
	in.Sync()
	fmt.Println("Done")
	return nil
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	// at this point, if it didn't exist, we'd know.
	return true
}

func mainOnce(name string) {
	now := time.Now().Format(*format)
	outname := fmt.Sprintf("%s.%s.gz", name, now)
	n := 0
	for fileExists(outname) {
		n += 1
		outname = fmt.Sprintf("%s.%s.%d.gz", name, now, n)
	}
	copyTruncate(name, outname)
}

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		mainOnce(arg)
	}
}
