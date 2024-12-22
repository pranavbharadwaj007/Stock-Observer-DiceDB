package operations

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "strconv"
    "strings"
    "sync"
    // "time"

    "stock-observer/internal/observer/implementations/stock"
)

// StockUpdate is what we’ll pass through the channel
type StockUpdate struct {
    Symbol   string
    NewPrice float64
}

func ProcessUpdates(ctx context.Context, wg *sync.WaitGroup, market *stock.StockMarket, updates chan StockUpdate) {
    defer wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case upd := <-updates:
            // update the stock price in DiceDB and notify
            market.SetStockPrice(ctx, upd.Symbol, upd.NewPrice)
        }
    }
}

// ReadInputs prompts user for multiple stock updates
func ReadInputs(ctx context.Context, wg *sync.WaitGroup, updates chan StockUpdate) {
    defer wg.Done()

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Print("Enter number of updates: ")
    if !scanner.Scan() {
        fmt.Println("Failed to read number of updates")
        return
    }
    numUpdates, err := strconv.Atoi(scanner.Text())
    if err != nil {
        fmt.Printf("Invalid number of updates: %v\n", err)
        return
    }

    for i := 0; i < numUpdates; i++ {
        select {
        case <-ctx.Done():
            return
        default:
            fmt.Printf("Update [%d/%d] => Enter <Symbol> <NewPrice>: ", i+1, numUpdates)
            if !scanner.Scan() {
                fmt.Println("Failed to read input")
                return
            }
            line := scanner.Text()
            parts := strings.Fields(line)
            if len(parts) != 2 {
                fmt.Println("Invalid format: Expected SYMBOL PRICE")
                return
            }

            symbol := parts[0]
            newPrice, err := strconv.ParseFloat(parts[1], 64)
            if err != nil {
                fmt.Printf("Failed to parse new price: %v\n", err)
                return
            }

            updates <- StockUpdate{
                Symbol:   symbol,
                NewPrice: newPrice,
            }
        }
    }
}
