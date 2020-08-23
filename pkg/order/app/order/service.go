package order

func NewService(repo Repository, productsRetriever ProductsRetriever) *Service {
	return &Service{repo, productsRetriever}
}

type Service struct {
	repo              Repository
	productsRetriever ProductsRetriever
}

type ProductsRetriever interface {
	OrderProducts(userID string) ([]Product, error)
	RestoreProducts(userID string, products []Product) error
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
		_ = s.productsRetriever.RestoreProducts(userID, products)
		return orderID, err
	}
	// TODO: call payment service or create some event

	return orderID, nil
}

func (s *Service) PayOrder(orderID string) error {
	o, err := s.repo.FindByID(ID(orderID))
	if err != nil {
		return err
	}

	o.Status = Completed
	return s.repo.Store(o)
}
