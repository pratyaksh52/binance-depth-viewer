To structure this Go project properly into separate files and packages, we need to adhere to standard Go practices and ensure that each file/package has a clear responsibility. Here’s how we can organize the project:

1. **Project Structure:**
   ```
   /orderbook-app
   ├── main.go
   ├── go.mod
   ├── go.sum
   ├── internal
   │   ├── orderbook
   │   │   ├── orderbook.go
   │   │   └── orderbook_test.go
   │   ├── renderer
   │   │   └── renderer.go
   │   └── utils
   │       └── utils.go
   └── pkg
       └── binance
           └── binance.go
   ```

2. **File Breakdown:**
   - **`main.go`**: Entry point of the application.
   - **`internal/orderbook/orderbook.go`**: Order book logic.
   - **`internal/renderer/renderer.go`**: Terminal rendering logic.
   - **`internal/utils/utils.go`**: Utility functions like error logging.
   - **`pkg/binance/binance.go`**: Binance API related structures and functions.

Here's how to refactor your code into this structure:

### `main.go`
```go
package main

import (
	"time"
	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
	"orderbook-app/internal/orderbook"
	"orderbook-app/internal/renderer"
	"orderbook-app/internal/utils"
	"orderbook-app/pkg/binance"
)

const wsendpoint = "wss://fstream.binance.com/stream?streams=btcusdt@depth"
const defaultBackgroundColor = termbox.ColorDefault

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	utils.LogError(err)

	var (
		ob     = orderbook.NewOrderBook()
		result binance.BinanceDepthResponse
	)

	go func() {
		for {
			err := conn.ReadJSON(&result)
			utils.LogError(err)
			ob.HandleDepthResponse(result.Data)
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
		renderer.RenderOrderBook(ob, 50, 2)
		time.Sleep(time.Millisecond * 32)
		termbox.Flush()
	}
}
```

### `internal/orderbook/orderbook.go`
```go
package orderbook

import (
	"strconv"
	"github.com/VictorLowther/btree"
	"orderbook-app/pkg/binance"
)

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

func byBestBid(a, b *OrderBookEntry) bool {
	return a.Price >= b.Price
}

func byBestAsk(a, b *OrderBookEntry) bool {
	return a.Price < b.Price
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

func (ob *OrderBook) HandleDepthResponse(res binance.BinanceDepthResult) {
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
```

### `internal/renderer/renderer.go`
```go
package renderer

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/mattn/go-runewidth"
	"orderbook-app/internal/orderbook"
)

const defaultBackgroundColor = termbox.ColorDefault

func RenderOrderBook(ob *orderbook.OrderBook, x, y int) {
	// Render Asks
	for i, ask := range ob.getNAsks(10) {
		if ask == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", ask.Price)
		size := fmt.Sprintf("%.4f", ask.Size)
		RenderText(x, y+i, price, termbox.ColorRed)
		RenderText(x+10, y+i, size, termbox.ColorCyan)
	}

	for i, bid := range ob.getNBids(10) {
		if bid == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", bid.Price)
		size := fmt.Sprintf("%.4f", bid.Size)
		RenderText(x, y+i+10, price, termbox.ColorGreen)
		RenderText(x+10, y+i+10, size, termbox.ColorCyan)
	}
}

func RenderText(x, y int, msg string, color termbox.Attribute) {
	for _, ch := range msg {
		termbox.SetCell(x, y, ch, color, defaultBackgroundColor)
		width := runewidth.RuneWidth(ch)
		x += width
	}
}
```

### `internal/utils/utils.go`
```go
package utils

import (
	"log"
)

func LogError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
```

### `pkg/binance/binance.go`
```go
package binance

type BinanceDepthResult struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}

type BinanceDepthResponse struct {
	Stream string             `json:"stream"`
	Data   BinanceDepthResult `json:"data"`
}
```

### Additional Steps:
1. **Testing**: Write unit tests for each package in the `internal/orderbook/orderbook_test.go` file.
2. **Dependencies**: Make sure to run `go mod tidy` to clean up dependencies.

### Usage:
To build and run the application, navigate to the project root and use:
```sh
go build -o orderbook-app ./...
./orderbook-app
```

This structure enhances readability, maintainability, and allows for easier testing and modification.