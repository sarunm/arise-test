package transaction

import "context"

type Service interface {
	Deposit(ctx context.Context, req DepositRequest) (*TransactionResponse, error)
	Withdraw(ctx context.Context, req WithdrawRequest) (*TransactionResponse, error)
	Transfer(ctx context.Context, req TransferRequest) (*TransactionResponse, error)
	GetByAccountID(ctx context.Context, accountID int) ([]TransactionResponse, error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) Deposit(ctx context.Context, req DepositRequest) (*TransactionResponse, error) {
	tx, err := s.repo.Deposit(ctx, req)
	if err != nil {
		return nil, err
	}
	res := tx.ToResponse()
	return &res, nil
}

func (s *service) Withdraw(ctx context.Context, req WithdrawRequest) (*TransactionResponse, error) {
	tx, err := s.repo.Withdraw(ctx, req)
	if err != nil {
		return nil, err
	}
	res := tx.ToResponse()
	return &res, nil
}

func (s *service) Transfer(ctx context.Context, req TransferRequest) (*TransactionResponse, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}
	tx, err := s.repo.Transfer(ctx, req)
	if err != nil {
		return nil, err
	}
	res := tx.ToResponse()
	return &res, nil
}

func (s *service) GetByAccountID(ctx context.Context, accountID int) ([]TransactionResponse, error) {
	txs, err := s.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	responses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = tx.ToResponse()
	}
	return responses, nil
}
