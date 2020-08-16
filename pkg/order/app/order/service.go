package order

func NewService(repo Repository, productsRetriever OrderProductsRetriever) *Service {
	return &Service{repo, productsRetriever}
}

type Service struct {
	repo              Repository
	productsRetriever OrderProductsRetriever
}

type OrderProductsRetriever interface {
	OrderProducts(userID string) ([]Product, error)
}

func (s *Service) CreateOrder(userID string) (orderID ID, err error) {
	orderID, err = s.repo.NextID()
	if err != nil {
		return orderID, err
	}

	products, err := s.productsRetriever.OrderProducts(userID)
	if err != nil {
		return orderID, err
	}

	if len(products) == 0 {
		return orderID, ErrEmptyCart
	}

	o := &Order{
		ID:       orderID,
		UserID:   userID,
		Status:   PendingPayment,
		Products: products,
	}
	err = s.repo.Store(o)
	if err != nil {
		return orderID, err
	}
	// TODO: call payment service or create some event

	return orderID, nil
}
