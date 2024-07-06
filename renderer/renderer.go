package renderer

import (
	"fmt"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"github.com/pratyaksh52/binance-depth-viewer/orderbook"
)

const defaultBackgroundColor = termbox.ColorDefault
const defaultLimit = 10

func RenderOrderBook(ob *orderbook.OrderBook, x, y int) {
	// Render Asks
	for i, ask := range ob.GetNAsks(defaultLimit) {
		if ask == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", ask.Price)
		size := fmt.Sprintf("%.4f", ask.Size)
		RenderText(x, y+i, price, termbox.ColorRed)
		RenderText(x+10, y+i, size, termbox.ColorCyan)
	}

	for i, bid := range ob.GetNBids(defaultLimit) {
		if bid == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", bid.Price)
		size := fmt.Sprintf("%.4f", bid.Size)
		RenderText(x, y+i+defaultLimit, price, termbox.ColorGreen)
		RenderText(x+defaultLimit, y+i+defaultLimit, size, termbox.ColorCyan)
	}
}

func RenderText(x, y int, msg string, color termbox.Attribute) {
	for _, ch := range msg {
		termbox.SetCell(x, y, ch, color, defaultBackgroundColor)
		width := runewidth.RuneWidth(ch)
		x += width
	}
}
