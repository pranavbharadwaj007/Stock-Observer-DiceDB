package operations

import (
    "context"
    "fmt"
    "math/rand"
    "strconv"
    "time"
    "github.com/dicedb/dicedb-go"
    "stock-observer/internal/observer/implementations/investor"
    "stock-observer/internal/observer/implementations/stock"
)

// CreateInvestorFlow prompts user for data, creates an investor, and stores it.
func CreateInvestorFlow(ctx context.Context, db *dicedb.Client) error {
    var (
        name string
        threshold float64
        investorType string
    )
    fmt.Println("Enter investor name:")
    fmt.Scanln(&name)
    fmt.Println("Enter threshold:")
    fmt.Scanln(&threshold)
    fmt.Println("Enter investor type:")
    fmt.Scanln(&investorType)
    ID := GenerateInvestorID()

    inv := investor.Investor{
        ID: ID,
        Name: name,
        Threshold: threshold,
        Type: investorType,
    }
    if err := StoreInvestorInDB(ctx, db, inv); err != nil {
        fmt.Printf("Failed to store investor: %v\n", err)
        return err
    }
    // Add the investor id to the set of investors
    if err := db.SAdd(ctx, "investors", ID).Err(); err != nil {
        fmt.Printf("Failed to add investor to set: %v\n", err)
        return err
    }
    fmt.Printf("Investor %s created successfully\n", ID)
    return nil
}

// ListInvestors shows all investor IDs, fetches details, prints them
func ListInvestors(ctx context.Context, db *dicedb.Client) error {
    // Get all investor IDs
    ids, err := db.SMembers(ctx, "investors").Result()
    if err != nil {
        fmt.Printf("Failed to get investor IDs: %v\n", err)
        return err
    }
    if len(ids) == 0 {
        fmt.Println("No investors found")
        return nil
    }
    fmt.Println("-------------------------------------------------")
    fmt.Printf("%-10s | %-15s | %-10s | %-10s\n", "ID", "Name", "Threshold", "Type")
    fmt.Println("-------------------------------------------------")
    for _, id := range ids {
        inv, err := LoadInvestorByID(ctx, db, id)
        if err != nil {
            fmt.Printf("Failed to load investor %s: %v\n", id, err)
            continue
        }
        fmt.Printf("%-10s | %-15s | %-10.2f | %-10s\n", inv.ID, inv.Name, inv.Threshold, inv.Type)
    }
    return nil
}

func LoadInvestorByID(ctx context.Context, db *dicedb.Client, id string) (investor.Investor, error) {
    inv := investor.Investor{}
    key := "investor:" + id
    data, err := db.HGetAll(ctx, key).Result()
    if err != nil {
        return inv, err
    }
    if len(data) == 0 {
        return inv, fmt.Errorf("investor %s not found", id)
    }
    inv.ID = data["id"]
    inv.Name = data["name"]
    inv.Threshold, _ = strconv.ParseFloat(data["threshold"], 64)
    inv.Type = data["type"]
    return inv, nil
}

func SubscribeInvestorFlow(ctx context.Context, db *dicedb.Client, market *stock.StockMarket) error {
    // List all investors
    if err := ListInvestors(ctx, db); err != nil {
        fmt.Printf("Failed to list investors: %v\n", err)
        return err
    }
    // Prompt user for investor ID
    var invID string
    fmt.Println("Enter investor ID to subscribe:")
    fmt.Scanln(&invID)

    inv, err := LoadInvestorByID(ctx, db, invID)
    if err != nil {
        fmt.Printf("Failed to load investor %s: %v\n", invID, err)
        return err
    }
    // Subscribe investor to stock updates
    if err := market.RegisterObserver(ctx,&inv); err != nil {
        fmt.Printf("Failed to subscribe investor: %v\n", err)
        return err
    }
    fmt.Printf("Investor %s subscribed successfully\n", invID)
    return nil
}

func UnsubscribeInvestorFlow(ctx context.Context, db *dicedb.Client, market *stock.StockMarket) error {
	// List all investors
	if err := ListInvestors(ctx, db); err != nil {
		fmt.Printf("Failed to list investors: %v\n", err)
		return err
	}
	// Prompt user for investor ID
	var invID string
	fmt.Println("Enter investor ID to unsubscribe:")
	fmt.Scanln(&invID)

	inv, err := LoadInvestorByID(ctx, db, invID)
	if err != nil {
		fmt.Printf("Failed to load investor %s: %v\n", invID, err)
		return err
	}
	// Unsubscribe investor from stock updates
	if err := market.RemoveObserver(ctx,&inv); err != nil {
		fmt.Printf("Failed to unsubscribe investor: %v\n", err)
		return err
	}
	fmt.Printf("Investor %s unsubscribed successfully\n", invID)
	return nil
}

func StoreInvestorInDB(ctx context.Context, db *dicedb.Client, inv investor.Investor) error {
    key := "investor:" + inv.ID
    err := db.HSet(ctx, key, map[string]interface{}{
        "id":        inv.ID,
        "name":      inv.Name,
        "threshold": inv.Threshold,
        "type":      inv.Type,
    }).Err()
    return err
}

func GenerateInvestorID() string {
    rand.Seed(time.Now().UnixNano())
    return fmt.Sprintf("%06d", 100000+rand.Intn(900000))
}