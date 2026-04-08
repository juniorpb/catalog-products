package domain

// Product represents a catalog product stored in the database.
type Product struct {
	Id       int64
	Name     string
	Brand    string
	Category string
}

// Seller represents a marketplace seller.
type Seller struct {
	Id   string
	Name string
}

// SellerProduct links a seller to a catalog product, storing the seller's
// original product identifier as ExternalId.
type SellerProduct struct {
	SellerId   string
	ProductId  int64
	ExternalId string
}
