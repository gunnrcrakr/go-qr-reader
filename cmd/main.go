package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	logs "github.com/labstack/gommon/log"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

var FILE_PATH string = "./downloads/"

type Img struct {
	ImgURL string `json:"img_url"`
}

func main() {
	// Setup
	e := echo.New()
	e.Logger.SetLevel(logs.INFO)
	e.POST("/decode", process)

	go func() {
		if err := e.Start(":" + os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

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

	full_path, err := download(img.ImgURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"err": err.Error()})
	}

	qr_string, err := decode(full_path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"err": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"qr_string": qr_string.GetText()})
}

func download(URL string) (string, error) {

	response, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", response.StatusCode, response.Status)
	}

	file_name := uuid.New().String() + ".jpg"
	full_path := FILE_PATH + file_name
	file, err := os.Create(full_path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return full_path, nil
}

func decode(full_path string) (*gozxing.Result, error) {

	file, err := os.Open(full_path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}

	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
