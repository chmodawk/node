/*
 * Copyright (C) 2020 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package pilvytis

import (
	"fmt"
	"time"

	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/requests"
)

// API is object which exposes pilvytis API.
type API struct {
	req    *requests.HTTPClient
	signer identity.SignerFactory
	dsn    string
}

const orderEndpoint = "payment/orders"

// NewAPI returns a new API instance.
func NewAPI(hc *requests.HTTPClient, dsn string, signer identity.SignerFactory) *API {
	return &API{
		req:    hc,
		signer: signer,
		dsn:    dsn,
	}
}

// OrderResp is returned from the pilvytis order endpoints.
type OrderResp struct {
	ID       uint64 `json:"id"`
	Status   string `json:"status"`
	Identity string `json:"identity"`

	PriceAmount   *float64 `json:"price_amount"`
	PriceCurrency string   `json:"price_currency"`

	PayAmount      *float64 `json:"pay_amount,omitempty"`
	PayCurrency    *string  `json:"pay_currency,omitempty"`
	PaymentAddress string   `json:"payment_address"`

	ReceiveAmount   *float64 `json:"receive_amount"`
	ReceiveCurrency string   `json:"receive_currency"`

	ExpiresAt time.Time `json:"expire_at"`
	CreatedAt time.Time `json:"created_at"`
}

type orderReq struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	ReceiveCurrency  string  `json:"receive_currency"`
	LightningNetwork bool    `json:"lightning_network"`
}

// CreatePaymentOrder creates a new payment order in the API service.
func (a *API) CreatePaymentOrder(i identity.Identity, priceAmount float64, priceCurr, recvCurr string, ln bool) (OrderResp, error) {
	payload := orderReq{
		PriceCurrency:    priceCurr,
		PriceAmount:      priceAmount,
		ReceiveCurrency:  recvCurr,
		LightningNetwork: ln,
	}

	req, err := requests.NewSignedPostRequest(a.dsn, orderEndpoint, payload, a.signer(i))
	if err != nil {
		return OrderResp{}, err
	}

	var resp OrderResp
	return resp, a.req.DoRequestAndParseResponse(req, &resp)
}

// GetPaymentOrder returns a payment order by ID from the API
// service that belongs to a given identity.
func (a *API) GetPaymentOrder(i identity.Identity, oid uint64) (OrderResp, error) {
	req, err := requests.NewSignedGetRequest(a.dsn, fmt.Sprintf("%s/%d", orderEndpoint, oid), a.signer(i))
	if err != nil {
		return OrderResp{}, err
	}

	var resp OrderResp
	return resp, a.req.DoRequestAndParseResponse(req, &resp)
}

// GetPaymentOrders returns a list of payment orders from the API service made by a given identity.
func (a *API) GetPaymentOrders(i identity.Identity) ([]OrderResp, error) {
	req, err := requests.NewSignedGetRequest(a.dsn, orderEndpoint, a.signer(i))
	if err != nil {
		return nil, err
	}

	var resp []OrderResp
	return resp, a.req.DoRequestAndParseResponse(req, &resp)
}
