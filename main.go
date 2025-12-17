package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

func main() {

	// CLI arg check
	if len(os.Args) < 2 || len(os.Args) > 2 {
		fmt.Printf("the program need 1 argument but %d found.\n", len(os.Args)-1)
		fmt.Println("usage: go run main.go <url>")
		os.Exit(1)
	}

	targetURL := os.Args[1] //Args[0] -> main.go , Args[1] -> targetURL
	fmt.Println("Target URL:", targetURL)

	// create output folder
	outDir, childDirName, err := createOutputDirectory(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	htmlPath := filepath.Join(outDir, childDirName+".html")
	imgPath := filepath.Join(outDir, childDirName+".png")

	// fetch html
	err = fetchHTML(targetURL, htmlPath)
	if err != nil {
		log.Fatal("HTML Fetch Err:", err)
	}

	// screenshot
	err = takeScreenShot(targetURL, imgPath)
	if err != nil {
		log.Fatal("Screenshot Err:", err)
	}

	fmt.Println("Successfully completed.")
}

func fetchHTML(url string, outputPath string) error {

	c := colly.NewCollector(
		colly.UserAgent(
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "+
				"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.AllowURLRevisit(),
	)

	c.SetRequestTimeout(15 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Fetching HTML:", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		err := os.WriteFile(outputPath, r.Body, 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("HTML Saved:", outputPath)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Colly Err:", err)
	})

	return c.Visit(url)
}

func takeScreenShot(url string, outputPath string) error {

	start := time.Now()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var buf []byte

	fmt.Println("Taking Screenshot...")

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(5*time.Second),
		//chromedp.WaitVisible("body"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(3*time.Second),
		chromedp.FullScreenshot(&buf, 90),
		//chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, buf, 0644)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	fmt.Printf("Screenshot Saved: %s (in %s)\n", outputPath, elapsed)
	return nil
}

func createOutputDirectory(targetURL string) (string, string, error) {

	u, err := url.Parse(targetURL)
	if err != nil {
		return "", "", err
	}

	site := strings.ReplaceAll(u.Hostname(), "www.", "")
	site = strings.ReplaceAll(site, ".", "_")

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	childDirName := fmt.Sprintf("%s_%s", timestamp, site)

	baseDir := "outputs"
	fullDir := filepath.Join(baseDir, childDirName)

	err = os.MkdirAll(fullDir, 0755)
	if err != nil {
		return "", "", err
	}

	return fullDir, childDirName, nil
}
