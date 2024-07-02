package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func byBestBid(a, b *OrderBookEntry) bool {
	return a.Price >= b.Price
}

func byBestAsk(a, b *OrderBookEntry) bool {
	return a.Price < b.Price
}

type OrderBookEntry struct {
	Price float64
	Size  float64
}

type OrderBook struct {
	Asks *btree.Tree[*OrderBookEntry]
	Bids *btree.Tree[*OrderBookEntry]
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks: btree.New(byBestAsk),
		Bids: btree.New(byBestBid),
	}
}

func getAskByPrice(price float64) btree.CompareAgainst[*OrderBookEntry] {
	return func(e *OrderBookEntry) int {
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

func getBidsByPrice(price float64) btree.CompareAgainst[*OrderBookEntry] {
	return func(e *OrderBookEntry) int {
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

func (ob *OrderBook) handleDepthResponse(res BinanceDepthResult) {
	for _, ask := range res.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		size, _ := strconv.ParseFloat(ask[1], 64)

		if size == 0 {
			if event, ok := ob.Asks.Get(getAskByPrice(price)); ok {
				// fmt.Printf("Deleting entry %.2f", price)
				ob.Asks.Delete(event)
			}
		}

		entry := &OrderBookEntry{
			Price: price,
			Size:  size,
		}

		ob.Asks.Insert(entry)
	}
	for _, bid := range res.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		size, _ := strconv.ParseFloat(bid[1], 64)

		if size == 0 {
			if event, ok := ob.Bids.Get(getBidsByPrice(price)); ok {
				// fmt.Printf("Deleting entry %.2f", price)
				ob.Bids.Delete(event)
			}
		}

		entry := &OrderBookEntry{
			Price: price,
			Size:  size,
		}

		ob.Bids.Insert(entry)
	}
}

type BinanceDepthResult struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}

type BinanceDepthResponse struct {
	Stream string             `json:"stream"`
	Data   BinanceDepthResult `json:"data"`
}

const wsendpoint = "wss://fstream.binance.com/stream?streams=btcusdt@depth"

func (ob *OrderBook) render(x, y int) {
	// Render Ask
	itr := ob.Asks.Iterator(nil, nil)
	i := 0
	var item *OrderBookEntry
	var priceStr string
	for itr.Next() {
		item = itr.Item()
		priceStr := fmt.Sprintf("%.2f", item.Price)
		renderText(x, y+i, priceStr, termbox.ColorRed)
		i++
	}

	// Render Bid
	itr = ob.Bids.Iterator(nil, nil)
	i = 0
	x = x + 10
	for itr.Next() {
		item = itr.Item()
		priceStr = fmt.Sprintf("%.2f", item.Price)
		renderText(x, y+i, priceStr, termbox.ColorGreen)
		i++
	}
}

func renderText(x, y int, msg string, color termbox.Attribute) {
	for _, ch := range msg {
		termbox.SetCell(x, y, ch, color, termbox.ColorDefault)
		width := runewidth.RuneWidth(ch)
		x += width
	}
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	logError(err)

	var (
		ob     = NewOrderBook()
		result BinanceDepthResponse
	)

	go func() {
		for {
			err := conn.ReadJSON(&result)
			logError(err)
			ob.handleDepthResponse(result.Data)
		}
	}()

	termbox.Init()
	defer termbox.Close()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	isRunning := true
	go func() {
		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeySpace:
				ob.render(0, 0)
			case termbox.KeyEsc:
				isRunning = false
			}
		}
		time.Sleep(10 * time.Second)
		isRunning = false
	}()

	for isRunning {
		// event := termbox.PollEent()
		// case termbox.even
		ob.render(0, 0)
		termbox.Flush()
		// time.Sleep(2 * time.Second)
	}
}

// func main() {
// 	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
// 	logError(err)

// 	var (
// 		ob     = NewOrderBook()
// 		result BinanceDepthResponse
// 	)
// 	for {
// 		err := conn.ReadJSON(&result)
// 		logError(err)

// 		ob.handleDepthResponse(result.Data)
// 		itr := ob.Bids.Iterator(nil, nil)
// 		for itr.Next() {
// 			item := itr.Item()
// 			fmt.Printf("%+v\n", item)
// 		}
// 	}
// }
