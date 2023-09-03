package main

import (
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"os"
	"strings"
	"time"
)

type guide struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type category struct {
	Url string
}

func main() {
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.OnError(func(resp *colly.Response, err error) {
		resp.Request.Retry()
	})
	c.Limit(&colly.LimitRule{
		Delay: 1 * time.Second,
	})
	meilisearchClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: "http://127.0.0.1:7700",
	})

	guides := scrape(c)
	addToMeilisearch(guides, meilisearchClient)
}

func scrape(c *colly.Collector) []guide {
	categories := getCategories(c)
	guides := []guide{}

	c.OnHTML(".entry-content > ol", func(h *colly.HTMLElement) {
		h.ForEach("li", func(_ int, el *colly.HTMLElement) {
			guides = append(guides, guide{
				Id:   uuid.NewString(),
				Name: strings.Title(strings.ToLower(el.ChildText("a"))),
				Url:  strings.ReplaceAll(el.ChildAttr("a", "href"), "http://mz.gov.mk", "https://zdravstvo.gov.mk"),
			})
		})
	})

	for _, url := range categories {
		c.Visit(url.Url)
	}

	c.Wait()

	return guides
}

func getCategories(c *colly.Collector) []category {
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

func addToMeilisearch(guides []guide, meilisearchClient *meilisearch.Client) {
	evidenceIndex := meilisearchClient.Index("evidence-based-medicine")

	_, err := evidenceIndex.AddDocuments(guides)
	if err != nil {
		os.Exit(1)
	}
}
