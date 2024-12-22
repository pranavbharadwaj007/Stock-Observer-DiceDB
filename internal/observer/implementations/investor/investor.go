package investor

import "fmt"

type Investor struct {
	ID string
	Name string
	Threshold float64
	Type string
}

func (i *Investor) Update(stockSymbol string, newPrice float64) {
	fmt.Printf("Investor %s received a notification that the price of %s has changed to %.2f\n", i.Name, stockSymbol, newPrice)
}

func (inv *Investor) GetID() string {
    return inv.ID
}
