package interfaces

type Observer interface {
	Update(stockSymbol string, newPrice float64)
	GetID() string
}
