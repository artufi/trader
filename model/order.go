package model

import (
	"database/sql"
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"time"
)

type Order struct {
	// PRIMARY KEY
	ID int
	// xtb order no
	Number int
	// xtb position, null indicates position get failure
	Position *int
	UserID   int
	// from PredictionDetails
	PredictionID int

	// from TradeTransactionStatus
	RequestStatus string
	Message       *string

	command.TradeTransInfo

	OrderClosedDetails
}

type OrderService struct {
	DB *sql.DB
}

func (os OrderService) SelectOpenOrdersWithPosition() ([]Order, error) {
	// now() returns UTC timezone, created_at has timezone from application
	// and need to be converted into UTC
	rows, err := os.DB.Query(`
				SELECT 
				    o.order_no,
				    o.position,
				    o.symbol,
				    o.volume,
				    o.closed
				FROM orders o
				JOIN predictions p ON o.prediction_id = p.id
                WHERE 
                	o.position IS NOT NULL 
                  	AND o.closed = false
                  	AND (now() - p.interval::INTERVAL) > (o.created_at AT TIME ZONE 'Universal')`)
	if err != nil {
		return nil, fmt.Errorf("select open orders with positions: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0, 0)
	for rows.Next() {
		order := Order{}
		err = rows.Scan(&order.Number, &order.Position, &order.Symbol, &order.Volume, &order.Closed)
		if err != nil {
			return nil, fmt.Errorf("select open orders with positions: %w", err)
		}
		orders = append(orders, order)
	}
	//err = rows.Err()

	return orders, nil
}

// get position to update order record?
func (os OrderService) SelectOpenOrders() ([]Order, error) {
	rows, err := os.DB.Query(`
				SELECT 
				    o.order_no,
				    o.position,
				    o.symbol,
				    o.volume,
				    o.closed
				FROM orders o
				JOIN predictions p ON o.prediction_id = p.id
                WHERE
                  	o.closed = false
                  	AND (now() - p.interval::INTERVAL) > (o.created_at AT TIME ZONE 'Universal')`)
	if err != nil {
		return nil, fmt.Errorf("select open orders: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0, 0)
	for rows.Next() {
		order := Order{}
		err = rows.Scan(&order.Number, &order.Position, &order.Symbol, &order.Volume, &order.Closed)
		if err != nil {
			return nil, fmt.Errorf("select open orders: %w", err)
		}
		if order.Number != 0 {
			orders = append(orders, order)
		}
	}
	//err = rows.Err()

	return orders, nil
}

func (os OrderService) SelectClosedOrdersById(orderID int) ([]Order, error) {
	rows, err := os.DB.Query(`
				SELECT 
				    o.order_no,
				    o.position,
				    o.symbol,
				    o.volume,
				    o.closed
				FROM orders o
				JOIN predictions p ON o.prediction_id = p.id
                WHERE
                    o.order_no = $1
                  	AND o.closed = true
                  	AND o.close_time IS NULL
                  	AND o.close_price IS NULL
                  	AND o.tts_message IS NULL`, orderID)
	if err != nil {
		return nil, fmt.Errorf("select closed orders by id: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0, 0)
	for rows.Next() {
		order := Order{}
		err = rows.Scan(&order.Number, &order.Position, &order.Symbol, &order.Volume, &order.Closed)
		if err != nil {
			return nil, fmt.Errorf("select closed orders by id: %w", err)
		}
		orders = append(orders, order)
	}
	//err = rows.Err()

	return orders, nil
}

func (os OrderService) Insert(o Order) (int, error) {
	row := os.DB.QueryRow(`
		INSERT INTO orders
			(order_no,
			 position,
			 user_id,
			 prediction_id,
			 symbol,
			 operation_code_cmd,
			 transaction_type,
			 custom_comment,
			 expiration,
			 offst,
			 ordr,
			 price,
			 stop_loss,
			 take_profit,
			 volume,
			 tts_request_status,
			 tts_message,
			 open_price,
			 close_price,
			 profit,
			 closed,
			 comment,
			 open_time,
			 close_time,
			 created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, 
		        $23, $24, $25)
		RETURNING id`, o.Number, o.Position, o.UserID, o.PredictionID, o.Symbol, o.Cmd, o.Type, o.CustomComment, o.Expiration,
		o.Offset, o.Order, o.Price, o.Sl, o.Tp, o.Volume, o.RequestStatus, o.Message, o.OpenPrice, o.ClosePrice,
		o.Profit, o.Closed, o.Comment, o.OpenTime, o.CloseTime, time.Now())

	err := row.Scan(&o.ID)
	if err != nil {
		return o.ID, fmt.Errorf("insert order: %w", err)
	}
	return o.ID, nil
}

func (os OrderService) UpdateOpened(orderID int, openPrice float64, openTime time.Time) (int, error) {
	row := os.DB.QueryRow(`
		UPDATE orders
		SET 
		    open_price = $2,
		    open_time = $3
		WHERE order_no = $1
		RETURNING id`, orderID, openPrice, openTime)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("update opened order: %w", err)
	}
	return id, nil
}

type OrderClosedDetails struct {
	// store nulls in the database to avoid ambiguity between zero values
	OpenPrice  *float64
	ClosePrice *float64
	Profit     *float64
	Closed     bool
	Comment    *string
	OpenTime   *time.Time
	CloseTime  *time.Time
}

func (os OrderService) UpdateClosed(orderID int, details OrderClosedDetails) (int, error) {
	row := os.DB.QueryRow(`
		UPDATE orders
		SET 
		    open_price = $2,
		    close_price = $3,
		    profit = $4, 
		    closed = $5,
		    comment = $6,
		    open_time = $7,
		    close_time = $8
		WHERE order_no = $1
		RETURNING id`, orderID, details.OpenPrice, details.ClosePrice, details.Profit, details.Closed,
		details.Comment, details.OpenTime, details.CloseTime)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("update closed order: %w", err)
	}
	return id, nil
}

func (os OrderService) MarkFailedAsClosed(orderID int) (int, error) {
	row := os.DB.QueryRow(`
		UPDATE orders
		SET 
		    closed = true
		WHERE order_no = $1
		RETURNING id`, orderID)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("mark order: %w", err)
	}
	return id, nil
}
