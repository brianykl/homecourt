package producers

import (
	"log"

	"fmt"

	"github.com/playwright-community/playwright-go"
)

type Player struct {
	Name     string
	Number   string
	Picture  string
	Position string
}

var teams = []string{
	"hawks", "celtics", "nets",
}

func HandleRoster() {
	pw, err := playwright.Run()

	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	defer pw.Stop()
	log.Printf("started playwright")

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()
	log.Printf("launched browser")

	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("could not create browser context: %v", err)
	}
	defer context.Close()
	log.Printf("created browser context")

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	log.Printf("created page")

	team := "celtics"
	url := "https://www.nba.com/" + team + "/roster"

	// Navigate to the page
	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}); err != nil {
		log.Fatalf("could not navigate to %s: %v", url, err)
	}
	log.Printf("went to %s", url)

	// Parameters for DOM stability
	const (
		stabilityDuration = 2000  // milliseconds to consider the DOM stable
		checkInterval     = 500   // milliseconds between DOM checks
		maxTimeout        = 30000 // maximum total wait time in milliseconds
	)

	// JavaScript function to check DOM stability
	script := fmt.Sprintf(`
		() => {
			return new Promise((resolve, reject) => {
				const stabilityDuration = %d;
				const maxTimeout = %d;
				const checkInterval = %d;
				let lastHTMLSize = 0;
				let lastCheckTime = Date.now();
				let checkCount = 0;

				const checkDOM = () => {
					const htmlSize = document.body.innerHTML.length;
					if (htmlSize !== lastHTMLSize) {
						lastCheckTime = Date.now();
						lastHTMLSize = htmlSize;
					}

					if (Date.now() - lastCheckTime >= stabilityDuration) {
						resolve();
					} else if ((Date.now() - lastCheckTime) > maxTimeout) {
						reject('Timeout: DOM did not stabilize');
					} else {
						setTimeout(checkDOM, checkInterval);
					}
				};
				checkDOM();
			});
		}
	`, stabilityDuration, maxTimeout, checkInterval)

	// Wait for the DOM to stabilize
	if _, err := page.WaitForFunction(script, nil); err != nil {
		log.Fatalf("DOM did not stabilize: %v", err)
	}

	log.Println("DOM has stabilized; proceeding to capture content.")

	log.Printf("finished waiting for dynamic content")

	// Capture the HTML content
	if _, err := page.PDF(playwright.PagePdfOptions{
		Path: playwright.String("output.pdf"),
	}); err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	// html, err := page.Content()
	// if err != nil {
	// 	log.Fatalf("could not get page content: %v", err)
	// }

	// log.Println("HTML content fetched successfully!")
	// log.Println(html) // This is the raw HTML you can send to GPT

}

// Helper function to trim text content
func trimText(text string) string {
	return text
}
