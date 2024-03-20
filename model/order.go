package model

import (
	"database/sql"
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"time"
)

type Order struct {
	ID            int
	Number        int
	UserID        int
	RequestStatus string
	Message       string
	command.TradeTransInfo
}

type OrderService struct {
	DB *sql.DB
}

func (os OrderService) Insert(o Order) (int, error) {
	row := os.DB.QueryRow(`
		INSERT INTO orders
			(order_no,
			 user_id,
			 symbol,
			 operation_code_cmd,
			 transaction_type,
			 expiration,
			 price,
			 stop_loss,
			 take_profit,
			 volume,
			 request_status,
			 message,
			 created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`, o.Number, o.UserID, o.Symbol, o.Cmd, o.Type, o.Expiration, o.Price, o.Sl, o.Tp, o.Volume,
		o.RequestStatus, o.Message, time.Now())

	err := row.Scan(&o.ID)
	if err != nil {
		return o.ID, fmt.Errorf("insert order: %w", err)
	}
	return o.ID, nil
}
