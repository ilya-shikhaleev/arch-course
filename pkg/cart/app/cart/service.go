package cart

func NewService(repo Repository) *Service {
	return &Service{repo}
}

type Service struct {
	repo Repository
}

func (s *Service) AddProductToCart(userID, productID string) (err error) {
	c, err := s.repo.FindByUserID(userID)
	if err == ErrCartNotFound {
		id, err := s.repo.NextID()
		if err != nil {
			return err
		}
		c = &Cart{
			ID:     id,
			UserID: userID,
		}
	} else if err != nil {
		return err
	}
	c.ProductIDs = append(c.ProductIDs, productID)

	return s.repo.Store(c)
}
