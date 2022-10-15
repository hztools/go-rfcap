package main

import (
	"io"
	"log"
	"os"
	"time"

	"hz.tools/rfcap"
	"hz.tools/rfcap/uring"
	"hz.tools/sdr"
)

func ohshit(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fd, err := os.Open(os.Args[1])
	ohshit(err)
	defer fd.Close()

	ring, err := uring.NewRing()
	ohshit(err)

	reader, err := ring.Reader(fd, uring.ReaderOpts{
		IQLength: 1024 * 32,
	})
	ohshit(err)
	start := time.Now()
	for reader.Next() {
		continue
	}
	end := time.Now()
	dur := end.Sub(start)
	log.Printf("uring Duration: %s", dur)

	fd.Seek(0, 0)
	rdr, header, err := rfcap.Reader(fd)
	ohshit(err)

	buf, err := sdr.MakeSamples(header.SampleFormat, 32*1024)
	ohshit(err)

	start = time.Now()
	for {
		_, err := rdr.Read(buf)
		if err == io.EOF {
			break
		} else {
			ohshit(err)
		}
	}
	end = time.Now()
	dur = end.Sub(start)
	log.Printf("sdr.Reader Duration: %s", dur)
}
