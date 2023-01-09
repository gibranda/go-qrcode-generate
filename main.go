package main

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/xid"
	"github.com/skip2/go-qrcode"
)

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func downloadPNG(c echo.Context) error {
	total := c.Param("total")
	totalInt, _ := strconv.Atoi(total)
	count := totalInt
	path := "qrcode/"

	// Generate QRCode to PNG
	for i := 0; i < count; i++ {
		// generate qrcode
		id := xid.New().String()
		fileName := strconv.Itoa(i) + "-" + id + ".png"
		err := qrcode.WriteFile(strconv.Itoa(i)+"-"+id, qrcode.Medium, 256, path+fileName)

		if err != nil {
			panic(err)
		}

		// Save to database
	}

	// Wrap to zip files
	if err := zipSource("qrcode", "qrcode.zip"); err != nil {
		log.Fatal(err)
		return c.String(http.StatusOK, "Failed")
	}

	// remove files after process done
	defer os.Remove("qrcode.zip")
	defer RemoveGlob("./qrcode/*")

	return c.File("qrcode.zip")
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/download-png/:total", downloadPNG)

	// Start server
	e.Logger.Fatal(e.Start(":1717"))

	// Generate QRCode to SVG
	// qr, err := qrsvg.New(fileName)
	// if err != nil {
	// 	panic(err)
	// }

	// // qr satisfies fmt.Stringer interface (or call qr.String() for a string)
	// fmt.Println(qr.String())
}

func RemoveGlob(path string) (err error) {
	contents, err := filepath.Glob(path)
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			return
		}
	}
	return
}

func zipSource(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
