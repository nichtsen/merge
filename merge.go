package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

const (
	//MAXB max bytes
	MAXB int64 = 100 * 1024 * 1024
	//BE big endian
	BE int32 = '\ufeff'
	//LE little endian
	LE int32 = '\ufffe'
)

var (
	files = make(map[string]string)
	total int
)
var (
	//ErrExceed Exceed the size limitation
	ErrExceed = errors.New("Exceed the size limitation")
)

func main() {
	cfg, err := loadConfig("config")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chdir(cfg.Root)
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(".", func(_path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if _path == "." {
			return nil
		}
		ext := path.Ext(_path)
		if ext == cfg.EXT {
			size := info.Size()
			//TODO add limitation size to config file
			if size > MAXB {
				log.Println("exceed the max size 100MB, current: ", size, " Path: ", _path)
				return ErrExceed
			}
			files[path.Join(cfg.Root, _path)] = info.ModTime().Format("2006-01-02 15:04:05.999999999 -0700 MST")
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range files {
		fmt.Println(k, v)
	}
	err = os.Chdir(cfg.WD)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(cfg.Target + cfg.EXT)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Println(f.Name())

	for k, v := range files {
		err = merge(f, k, v)
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Println(err)
			break
		}
	}
	if err == io.EOF {
		log.Println("Successfully merge!")
	}
}

func merge(f *os.File, _path string, date string) error {
	fr, err := os.Open(_path)
	if err != nil {
		return err
	}
	defer fr.Close()

	_, err = f.WriteString(header(_path, date))
	if err != nil {
		return err
	}

	b2 := make([]byte, 2)
	fr.ReadAt(b2, 0)
	end := rune(binary.BigEndian.Uint16(b2))
	switch end {
	case BE:
		_, err = fr.Read(b2) //pre-read
		if err != nil {
			return err
		}
		err = writeB(f, fr)
		if err != nil {
			return err
		}
	case LE:
		_, err = fr.Read(b2)
		if err != nil {
			return err
		}
		err = writeL(f, fr)
		if err != nil {
			return err
		}
	default:
		err = write(f, fr)
		if err != nil {
			return err
		}
	}
	return nil
}

func write(f *os.File, fr *os.File) error {
	var (
		n   int
		err error
	)
	b := make([]byte, 1024)
	for {
		n, err = fr.Read(b)
		if err != nil {
			return err
		}
		total += n
		if total > int(MAXB) {
			return ErrExceed
		}
		_, err = f.Write(b[:n])
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
	}
}

func writeB(f *os.File, fr *os.File) error {
	var (
		n   int
		err error
		buf buffer
	)
	b := make([]byte, 1024)
	for {
		n, err = fr.Read(b)
		if err != nil {
			return err
		}
		total += n
		if total > int(MAXB) {
			return ErrExceed
		}
		buf = encodeB(b, n)
		_, err = f.Write(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		buf = buf[0:0]
	}
}

func writeL(f *os.File, fr *os.File) error {
	var (
		n   int
		err error
		buf buffer
	)
	b := make([]byte, 1024)
	for {
		n, err = fr.Read(b)
		if err != nil {
			return err
		}
		total += n
		if total > int(MAXB) {
			return ErrExceed
		}
		buf = encodeL(b, n)
		_, err = f.Write(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		buf = buf[0:0]
	}
}

func isnil(b []byte) bool {
	for _, v := range b {
		if v == 0 {
			return true
		}
	}
	return false
}

func stirpnil(d []byte) ([]byte, int) {
	strip := make([]byte, 1024)
	var i int
	for _, v := range d {
		if v != 0 {
			strip[i] = v
			i++
		}
	}
	return strip, i
}

func header(_path string, date string) string {
	return fmt.Sprintf("\n/****** Path: %s Modify Date: %s ******/\n", _path, date)
}

func encodeL(b []byte, n int) buffer {
	var buf buffer
	for i := 0; i < n; i += 2 {
		r := rune(binary.LittleEndian.Uint16(b[i:]))
		buf.WriteRune(r)
	}
	return buf
}

func encodeB(b []byte, n int) buffer {
	var buf buffer
	for i := 0; i < n; i += 2 {
		r := rune(binary.BigEndian.Uint16(b[i:]))
		buf.WriteRune(r)
	}
	return buf
}
