package orderbook

import (
	"strconv"

	"github.com/pratyaksh52/binance-depth-viewer/binance"

	"github.com/VictorLowther/btree"
)

type BookEntry struct {
	Price float64
	Size  float64
}

type OrderBook struct {
	Asks *btree.Tree[*BookEntry]
	Bids *btree.Tree[*BookEntry]
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks: btree.New(byBestAsk),
		Bids: btree.New(byBestBid),
	}
}

func byBestBid(a, b *BookEntry) bool {
	return a.Price >= b.Price
}

func byBestAsk(a, b *BookEntry) bool {
	return a.Price < b.Price
}

func getAskByPrice(price float64) btree.CompareAgainst[*BookEntry] {
	return func(e *BookEntry) int {
		switch {
		case e.Price < price:
			return -1
		case e.Price > price:
			return 1
		default:
			return 0
		}
	}
}

func getBidsByPrice(price float64) btree.CompareAgainst[*BookEntry] {
	return func(e *BookEntry) int {
		switch {
		case e.Price > price:
			return -1
		case e.Price < price:
			return 1
		default:
			return 0
		}
	}
}

func (ob *OrderBook) HandleDepthResponse(res binance.DepthResult) {
	for _, ask := range res.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		size, _ := strconv.ParseFloat(ask[1], 64)

		if entry, ok := ob.Asks.Get(getAskByPrice(price)); ok {
			if size == 0 {
				ob.Asks.Delete(entry)
			} else {
				entry.Size = size
			}
			continue
		}

		entry := &BookEntry{
			Price: price,
			Size:  size,
		}

		ob.Asks.Insert(entry)
	}
	for _, bid := range res.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		size, _ := strconv.ParseFloat(bid[1], 64)

		if entry, ok := ob.Bids.Get(getBidsByPrice(price)); ok {
			if size == 0 {
				ob.Bids.Delete(entry)
			} else {
				entry.Size = size
			}
			continue
		}

		entry := &BookEntry{
			Price: price,
			Size:  size,
		}

		ob.Bids.Insert(entry)
	}
}

func (ob *OrderBook) GetNBids(depth int) []*BookEntry {
	var (
		bids = make([]*BookEntry, depth)
		itr  = ob.Bids.Iterator(nil, nil)
		i    = 0
	)

	for itr.Next() {
		if i == depth {
			break
		}
		bids[i] = itr.Item()
		i++
	}
	return bids
}

func (ob *OrderBook) GetNAsks(depth int) []*BookEntry {
	var (
		asks = make([]*BookEntry, depth)
		itr  = ob.Asks.Iterator(nil, nil)
		i    = 0
	)

	for itr.Next() {
		if i == depth {
			break
		}
		asks[i] = itr.Item()
		i++
	}
	return asks
}
