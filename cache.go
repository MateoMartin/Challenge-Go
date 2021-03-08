package sample1

import (
	"fmt"
	"sync"
	"time"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             map[string]*PriceItem
	mu                 sync.Mutex
}

// PriceItem is the item stored in the cache with its creation date and its corresponding price.
type PriceItem struct {
	dateCreated *time.Time
	price       float64
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		prices:             map[string]*PriceItem{},
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	priceItem, ok := c.prices[itemCode]
	if ok {
		if time.Now().Before(priceItem.dateCreated.Add(c.maxAge)) {
			return priceItem.price, nil
		}
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	dateCreated := time.Now()
	priceItem = &PriceItem{dateCreated: &dateCreated, price: price}
	c.mu.Lock()
	c.prices[itemCode] = priceItem
	c.mu.Unlock()
	return price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	results := []float64{}
	priceChan := make(chan float64, len(itemCodes))
	errChan := make(chan error)
	for _, itemCode := range itemCodes {
		go func(itemCode string) {
			price, err := c.GetPriceFor(itemCode)
			if err != nil {
				errChan <- err
			}
			priceChan <- price
		}(itemCode)
	}

	for i := 0; i < len(itemCodes); i++ {
		select {
		case price := <-priceChan:
			results = append(results, price)
		case err := <-errChan:
			return []float64{}, err
		}
	}

	return results, nil
}
