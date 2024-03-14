package controller

import (
	"context"
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/processor"
	"github.com/artufi/trader/xtb/response"
)

type Handler struct {
	Proc    processor.Proc
	TraceID string
}

type RespChecker interface {
	CheckStatus() error
	CheckCustomTag(customTag string) error
}

func (h Handler) ValidateResp(resp RespChecker, customTag string) error {
	var err error
	if err = resp.CheckCustomTag(customTag); err != nil {
		return err
	}
	if err = resp.CheckStatus(); err != nil {
		return err
	}
	return nil
}

func (h Handler) Login(ctx context.Context, userID, password string) (response.General, error) {
	proc := "login"
	customTag := h.TraceID + proc

	resp, err := h.Proc.Login(ctx, customTag, userID, password)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}
	if err = h.ValidateResp(resp, customTag); err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}

	return resp, nil
}

func (h Handler) GetSymbolExtended(ctx context.Context, symbol string) (response.GetSymbolExtended, error) {
	proc := "getSymbolExtended"
	customTag := h.TraceID + proc + symbol

	resp, err := h.Proc.GetSymbolExtended(ctx, symbol, customTag)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}
	if err = h.ValidateResp(resp, customTag); err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}

	return resp, nil
}

func (h Handler) TradeTransaction(ctx context.Context, symbol string, tradeTransInfo command.TradeTransInfo) (response.TradeTransaction, error) {
	proc := "tradeTransaction"
	customTag := h.TraceID + proc + symbol

	resp, err := h.Proc.TradeTransaction(ctx, tradeTransInfo, customTag)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}
	if err = h.ValidateResp(resp, customTag); err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}

	return resp, nil
}

func (h Handler) TradeTransactionStatus(ctx context.Context, symbol string, orderNo int) (response.TradeTransactionStatus, error) {
	proc := "tradeTransactionStatus"
	customTag := h.TraceID + proc + symbol

	resp, err := h.Proc.TradeTransactionStatus(ctx, orderNo, customTag)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}
	if err = h.ValidateResp(resp, customTag); err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}

	return resp, err
}

func (h Handler) Logout(ctx context.Context) (response.General, error) {
	proc := "logout"
	customTag := h.TraceID + proc
	resp, err := h.Proc.Logout(ctx, customTag)
	if err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}
	if err = resp.CheckStatus(); err != nil {
		return resp, fmt.Errorf("%s: %w", proc, err)
	}

	return resp, nil
}
