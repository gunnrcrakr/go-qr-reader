package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	logs "github.com/labstack/gommon/log"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

var FILE_PATH string = "./downloads/qrcode.png"

type Img struct {
	ImgURL string `json:"img_url"`
}

func main() {
	// Setup
	e := echo.New()
	e.Logger.SetLevel(logs.INFO)
	e.POST("/decode", process)

	// Start server
	go func() {
		if err := e.Start(":" + os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func process(c echo.Context) error {

	var img Img
	if err := c.Bind(&img); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"err": err.Error()})
	}

	err := download(img.ImgURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"err": err.Error()})
	}

	qr_string, err := decode()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"err": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"qr_string": qr_string.GetText()})
}

func download(URL string) error {
	// Get img from url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", response.StatusCode, response.Status)
	}

	// Create empty file
	file, err := os.Create(FILE_PATH)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write body to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func decode() (*gozxing.Result, error) {
	// open and decode image file
	file, _ := os.Open(FILE_PATH)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// prepare BinaryBitmap
	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)

	return result, nil
}
