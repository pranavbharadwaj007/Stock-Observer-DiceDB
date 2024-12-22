package config

import (
    "fmt"
	"context"
    "github.com/dicedb/dicedb-go"
)

func NewDiceDBClient(ctx context.Context,host string, port int) (*dicedb.Client, error) {
    addr := fmt.Sprintf("%s:%d", host, port)
    client := dicedb.NewClient(&dicedb.Options{
        Addr: addr,
        // any additional settings (password, etc.)
    })
    // Optional: test the connection
    if _, err := client.Ping(ctx).Result(); err != nil {
        return nil, err
    }
    return client, nil
}
