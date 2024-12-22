package interfaces

type Subject interface {
	RegisterObserver(observer Observer) error
	RemoveObserver(observer Observer) error
	NotifyObservers(stockSymbol string, newPrice float64)
}