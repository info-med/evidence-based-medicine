package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"strings"
	"time"
)

type guide struct {
	Name string
	Url  string
}

type category struct {
	Url string
}

func main() {
	c := colly.NewCollector(
		colly.Async(true),
		colly.Debugger(&debug.LogDebugger{}),
	)
	c.OnError(func(resp *colly.Response, err error) {
		resp.Request.Retry()
	})
	c.Limit(&colly.LimitRule{
		Delay: 1 * time.Second,
	})

	scrape(c)
	// Add to Meilisearch
}

func scrape(c *colly.Collector) {
	fmt.Println("scrape()")
	categories := getCategories(c)
	guides := []guide{}

	c.OnHTML(".entry-content > ol", func(h *colly.HTMLElement) {
		h.ForEach("li", func(_ int, el *colly.HTMLElement) {
			guides = append(guides, guide{
				Name: strings.Title(strings.ToLower(el.ChildText("a"))),
				Url:  strings.ReplaceAll(el.ChildAttr("a", "href"), "http://mz.gov.mk", "https://zdravstvo.gov.mk"),
			})
		})
	})

	for _, url := range categories {
		fmt.Println("Visiting", url.Url)
		c.Visit(url.Url)
	}

	c.Wait()

	d, err := json.Marshal(guides)

	if err != nil {
		fmt.Printf("Error: %s", err)

		return
	}

	fmt.Println(string(d))
}

func getCategories(c *colly.Collector) []category {
	fmt.Println("getCategories()")
	categories := []category{}
	c.OnHTML(".entry-content > ul", func(h *colly.HTMLElement) {
		h.ForEach("li", func(_ int, el *colly.HTMLElement) {
			categories = append(categories, category{Url: strings.ReplaceAll(el.ChildAttr("a", "href"), "http://mz.gov.mk", "https://zdravstvo.gov.mk")})
		})
	})
	c.Visit("https://zdravstvo.gov.mk/upatstva_update/")
	c.Wait()

	return categories
}
