package stock

import (
    "math"
    "fmt"
    "sync"
    "errors"
    "context"
    "strings"
    "time"
    "strconv"
    "stock-observer/internal/observer/interfaces"
    "github.com/dicedb/dicedb-go"
)

type StockMarket struct {
    observers map[interfaces.Observer]struct{}
    mu        sync.RWMutex
    threshold float64

    dbClient  *dicedb.Client
}

func NewStockMarket(threshold float64, dbClient *dicedb.Client) *StockMarket {
    return &StockMarket{
        observers: make(map[interfaces.Observer]struct{}),
        threshold: threshold,
        dbClient:  dbClient,
    }
}

func (m *StockMarket) RegisterObserver(ctx context.Context, obs interfaces.Observer) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 1) Check if already exists in our in-memory map
    if _, exists := m.observers[obs]; exists {
        return errors.New("Observer already exists")
    }
    m.observers[obs] = struct{}{}

    // 2) Also store in DiceDB so we know which observers are subscribed
    if m.dbClient != nil {
        observerID := obs.GetID()
        err := m.dbClient.SAdd(ctx, "observers", observerID).Err() // or "stock:observers"
        if err != nil {
            fmt.Printf("Error storing observerID in DiceDB: %v\n", err)
            // Not returning an error here if you prefer “best effort”
        }
    }

    return nil
}

func (m *StockMarket) RemoveObserver(ctx context.Context, obs interfaces.Observer) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 1) Check if exists in our in-memory map
    if _, exists := m.observers[obs]; !exists {
        fmt.Println("Observer does not exist in in-memory map")
        // check if it exists in DiceDB and remove it
        if m.dbClient != nil {
            observerID := obs.GetID()
            err := m.dbClient.SRem(ctx, "observers", observerID).Err() // or "stock:observers"
            if err != nil {
                fmt.Printf("Error removing observerID from DiceDB: %v\n", err)
                // same note about whether to return the error or not
                // print all the oserver ids in the db
                ids, err := m.dbClient.SMembers(ctx, "observers").Result()
                if err != nil {
                    fmt.Printf("Failed to get observer IDs: %v\n", err)
                }
                fmt.Println("Observer IDs in DiceDB:")
                for _, id := range ids {
                    fmt.Println(id)
                }
                return err
            }
        }
    }


    // 2) Also remove from DiceDB
    if m.dbClient != nil {
        observerID := obs.GetID()
        err := m.dbClient.SRem(ctx, "observers", observerID).Err() // or "stock:observers"
        if err != nil {
            fmt.Printf("Error removing observerID from DiceDB: %v\n", err)
            // same note about whether to return the error or not
        }
    }

    return nil
}


func (m *StockMarket) NotifyObservers(symbol string, newPrice float64) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    for obs := range m.observers {
        obs.Update(symbol, newPrice)
    }
}

func (m *StockMarket) SetStockPrice(ctx context.Context,symbol string, newPrice float64) {
    // oldPrice is the price of the stock before the update if it exists else 0
    oldPrice, err := m.dbClient.Get(ctx, "stock:"+symbol).Float64()
    if err != nil {
        oldPrice = 0
    }

    priceChange := math.Abs(newPrice - oldPrice) / oldPrice * 100
    fmt.Printf("Price change: %.2f%%, Threshold: %.2f%%\n", priceChange, m.threshold)

    // 1) Save the new price to DiceDB
    if m.dbClient != nil {
        // Key: "stock:<symbol>", Value: <newPrice>
        err := m.dbClient.Set(ctx, "stock:"+symbol, fmt.Sprintf("%.2f", newPrice), 0).Err()
        if err != nil {
            fmt.Println("Error storing new price in DiceDB:", err)
        }

        // 2) Also store a historical record
        // E.g. LPUSH "history:<symbol>" "newPrice|timestamp"
        record := fmt.Sprintf("%.2f|%d", newPrice, time.Now().Unix())
        if err := m.dbClient.LPush(ctx,"history:"+symbol, record).Err(); err != nil {
            fmt.Println("Error pushing history to DiceDB:", err)
        }
    }

    if priceChange >= m.threshold {
        m.NotifyObservers(symbol, newPrice)
        // show the history
        m.GetStockHistory(ctx,symbol)

        
    }
}

func (m *StockMarket) GetStockHistory(ctx context.Context, symbol string) error {
    if m.dbClient == nil {
        return fmt.Errorf("dbClient is nil")
    }

    history, err := m.dbClient.LRange(ctx, "history:"+symbol, 0, -1).Result()
    if err != nil {
        return fmt.Errorf("error retrieving history: %v", err)
    }

    fmt.Printf("Stock history for %s:\n", symbol)
    fmt.Printf("%-10s | %-10s | %-25s\n", "Symbol", "Price", "Time (IST)")
    fmt.Println(strings.Repeat("-", 50))

    for _, entry := range history {
        parts := strings.Split(entry, "|")
        if len(parts) != 2 {
            continue
        }

        priceStr := parts[0]
        timestampStr := parts[1]
        ts, err := strconv.ParseInt(timestampStr, 10, 64)
        if err != nil {
            fmt.Printf("Error parsing timestamp: %v\n", err)
            continue
        }
        t := time.Unix(ts, 0)

        // Convert to IST
        istLocation, err := time.LoadLocation("Asia/Kolkata")
        if err != nil {
            fmt.Printf("Error loading Asia/Kolkata location: %v\n", err)
            continue
        }
        tIST := t.In(istLocation).Format("2006-01-02 15:04:05")
        fmt.Printf("%-10s | %-10s | %-25s\n", symbol, priceStr, tIST)
    }

    return nil
}

