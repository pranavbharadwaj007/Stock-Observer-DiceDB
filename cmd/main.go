package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "stock-observer/config"
    "stock-observer/internal/observer/implementations/stock"
    "stock-observer/internal/observer/operations"
    "sync"
    "syscall"
    "time"
)

func main() {
    // Connect to DiceDB
    host := "20.204.32.147"
    port := 8538
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Graceful shutdown on Ctrl+C
    stopChan := make(chan os.Signal, 1)
    signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-stopChan
        fmt.Println("\nReceived interrupt... shutting down.")
        cancel()
    }()

    dbClient, err := config.NewDiceDBClient(ctx, host, port)
    if err != nil {
        fmt.Printf("Error connecting to DiceDB: %v\n", err)
        return
    }

    // Create a stock market with a global threshold = 0 (or any default)
    stockMarket := stock.NewStockMarket(0, dbClient)

    for {
        if ctx.Err() != nil {
            // context canceled, break the loop
            break
        }

        fmt.Println("\n=================== STOCK OBSERVER ===================")
        fmt.Println("1) Create Investor")
        fmt.Println("2) Subscribe Investor")
        fmt.Println("3) Run Concurrent Stock Updates")
        fmt.Println("4) Unsubscribe Investor")
        fmt.Println("5) Exit")
        fmt.Print("Your choice: ")

        var choice int
        _, scanErr := fmt.Scan(&choice)
        if scanErr != nil {
            fmt.Println("Invalid input, try again.")
            continue
        }

        switch choice {
        case 1:
            // Create an investor
            operations.CreateInvestorFlow(ctx, dbClient)

        case 2:
            // Subscribe existing investor (list IDs, pick ID)
            operations.SubscribeInvestorFlow(ctx, dbClient, stockMarket)

        case 3:
            // Start concurrency for reading stock updates
            updates := make(chan operations.StockUpdate, 100)
            var wg sync.WaitGroup
            wg.Add(2)

            go operations.ProcessUpdates(ctx, &wg, stockMarket, updates)
            go operations.ReadInputs(ctx, &wg, updates)

            // After spawn, wait for them or time out
            go func() {
                wg.Wait()
                cancel() // signal main that we are done
            }()

            // Wait with 30s timeout
            select {
            case <-ctx.Done():
                fmt.Println("Done or user canceled.")
            case <-time.After(30 * time.Second):
                fmt.Println("Timeout reached for stock updates.")
                cancel()
            }
            wg.Wait()
        
        case 4:
            // Unsubscribe investor (list IDs, pick ID)
            operations.UnsubscribeInvestorFlow(ctx, dbClient, stockMarket)

        case 5:
            fmt.Println("Exiting...")
            return

        default:
            fmt.Println("Unknown choice, try again.")
        }
    }
}
