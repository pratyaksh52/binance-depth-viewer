package main

import (
	"time"

	"github.com/pratyaksh52/binance-indicator/binance"
	"github.com/pratyaksh52/binance-indicator/renderer"
	"github.com/pratyaksh52/binance-indicator/utils"

	"github.com/pratyaksh52/binance-indicator/orderbook"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
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
