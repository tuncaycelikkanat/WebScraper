package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

func main() {
	if err := run(); err != nil {
		log.Println("Program failed:", err)
		os.Exit(1)
	}
}

func run() error {

	if len(os.Args) != 2 {
		return fmt.Errorf("usage: go run main.go <url>")
	}

	targetURL := os.Args[1]

	// url normalization
	if !strings.HasPrefix(targetURL, "http://") &&
		!strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	fmt.Println("Target URL:", targetURL)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// output directory
	outDir, childDirName, err := createOutputDirectory(targetURL)
	if err != nil {
		return err
	}

	// check fetchings
	collyOK := false
	chromeOK := false

	defer func() {
		if !collyOK && !chromeOK {
			fmt.Println("Both Colly and Chromedp failed. There is no result saved.")
			_ = os.RemoveAll(outDir)
		}
	}()

	collyHTMLPath := filepath.Join(outDir, childDirName+"_colly.html")
	chromeHTMLPath := filepath.Join(outDir, childDirName+"_chromedp.html")
	imgPath := filepath.Join(outDir, childDirName+".png")

	// colly part
	fmt.Println("\n-> COLLY Trying To Fetch HTML")
	if err := fetchHTMLWithColly(targetURL, collyHTMLPath); err != nil {
		fmt.Println("Colly failed:", err)
	} else {
		collyOK = true
	}

	// chromedp part
	fmt.Println("\n-> CHROMEDP Trying To Fetch HTML And Taking Screenshot")
	if err := fetchWithChromedp(targetURL, chromeHTMLPath, imgPath, rng); err != nil {
		fmt.Println("Chromedp failed:", err)
	} else {
		chromeOK = true
	}

	fmt.Println("\nResults:")
	fmt.Println("Colly: ", collyOK)
	fmt.Println("Chrome: ", chromeOK)

	return nil
}

func fetchHTMLWithColly(targetURL, outputPath string) error {

	c := colly.NewCollector(
		colly.UserAgent(
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "+
				"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		),
		colly.AllowURLRevisit(),
	)

	c.SetRequestTimeout(15 * time.Second)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "tr-TR,tr;q=0.9,en-US;q=0.8")
		r.Headers.Set("Referer", "https://www.google.com/")
		fmt.Println("→ Request:", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("← Status:", r.StatusCode)
		fmt.Println("← Server:", r.Headers.Get("Server"))

		_ = os.WriteFile(outputPath, r.Body, 0644)
		fmt.Println("HTML saved:", outputPath)
	})

	c.OnError(func(r *colly.Response, err error) {
		if r != nil {
			fmt.Println("← Status:", r.StatusCode)
		}
	})

	return c.Visit(targetURL)
}

func fetchWithChromedp(targetURL, htmlPath, imgPath string, rng *rand.Rand) error {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-position", "-32000,-32000"),
		chromedp.Flag("window-size", "1920,1080"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var html string
	var screenshot []byte

	fmt.Println("Opening Real Browser, Please Just Wait...")

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(time.Duration(3+rng.Intn(3))*time.Second),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(time.Duration(2+rng.Intn(3))*time.Second),
		chromedp.OuterHTML("html", &html),
		chromedp.FullScreenshot(&screenshot, 90),
	)
	if err != nil {
		return err
	}

	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(imgPath, screenshot, 0644); err != nil {
		return err
	}

	fmt.Println("Chromedp HTML saved: ", htmlPath)
	fmt.Println("Screenshot saved: ", imgPath)

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

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", err
	}

	return fullDir, childDirName, nil
}
