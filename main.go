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

		if entry, ok := ob.Asks.Get(getAskByPrice(price)); ok {
			if size == 0 {
				ob.Asks.Delete(entry)
			} else {
				entry.Size = size
			}
			continue
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

		if entry, ok := ob.Bids.Get(getBidsByPrice(price)); ok {
			if size == 0 {
				ob.Bids.Delete(entry)
			} else {
				entry.Size = size
			}
			continue
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

func (ob *OrderBook) getNBids(depth int) []*OrderBookEntry {
	var (
		bids = make([]*OrderBookEntry, depth)
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

func (ob *OrderBook) getNAsks(depth int) []*OrderBookEntry {
	var (
		asks = make([]*OrderBookEntry, depth)
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

func (ob *OrderBook) render(x, y int) {
	// Render Asks
	for i, ask := range ob.getNAsks(10) {
		if ask == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", ask.Price)
		size := fmt.Sprintf("%.4f", ask.Size)
		renderText(x, y+i, price, termbox.ColorRed)
		renderText(x+10, y+i, size, termbox.ColorCyan)
	}

	for i, bid := range ob.getNBids(10) {
		if bid == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", bid.Price)
		size := fmt.Sprintf("%.4f", bid.Size)
		renderText(x, y+i+10, price, termbox.ColorGreen)
		renderText(x+10, y+i+10, size, termbox.ColorCyan)
	}
}

func renderText(x, y int, msg string, color termbox.Attribute) {
	for _, ch := range msg {
		termbox.SetCell(x, y, ch, color, defaultBackgroundColor)
		width := runewidth.RuneWidth(ch)
		x += width
	}
}

const wsendpoint = "wss://fstream.binance.com/stream?streams=btcusdt@depth"
const defaultBackgroundColor = termbox.ColorDefault

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

	isRunning := true
	eventch := make(chan termbox.Event, 1)
	defer close(eventch)
	go func() {
		for {
			eventch <- termbox.PollEvent()
		}
	}()

	for isRunning {
		termbox.Clear(defaultBackgroundColor, defaultBackgroundColor)
		select {
		case event := <-eventch:
			switch event.Key {
			case termbox.KeyEsc:
				isRunning = false
			default:
				continue
			}
		default:
		}
		ob.render(50, 2)
		time.Sleep(time.Millisecond * 32)
		termbox.Flush()
	}
}
