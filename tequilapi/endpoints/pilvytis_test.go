package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/pilvytis"
	"github.com/mysteriumnetwork/node/tequilapi/contract"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockPilvytis struct {
	identity string
	resp     pilvytis.OrderResp
}

func (mock *mockPilvytis) CreatePaymentOrder(i identity.Identity, priceAmount float64, priceCurr, recvCurr string, ln bool) (pilvytis.OrderResp, error) {
	if i.Address != mock.identity {
		return pilvytis.OrderResp{}, errors.New("wrong identity")
	}

	return mock.resp, nil
}

func (mock *mockPilvytis) GetPaymentOrder(i identity.Identity, oid uint64) (pilvytis.OrderResp, error) {
	if i.Address != mock.identity {
		return pilvytis.OrderResp{}, errors.New("wrong identity")
	}

	return mock.resp, nil
}

func (mock *mockPilvytis) GetPaymentOrders(i identity.Identity) ([]pilvytis.OrderResp, error) {
	if i.Address != mock.identity {
		return nil, errors.New("wrong identity")
	}

	return []pilvytis.OrderResp{mock.resp}, nil
}

func newMockPilvytisResp(id int, identity, priceC, recC string, recvAmount float64) pilvytis.OrderResp {
	s := "test"
	f := 1.0
	return pilvytis.OrderResp{
		ID:              uint64(id),
		Status:          "pending",
		Identity:        identity,
		PriceAmount:     &recvAmount,
		PriceCurrency:   priceC,
		PayAmount:       &f,
		PayCurrency:     &s,
		PaymentAddress:  "0x00",
		ReceiveAmount:   &f,
		ReceiveCurrency: recC,
		ExpiresAt:       time.Now(),
		CreatedAt:       time.Now(),
	}
}

func TestCreatePaymentOrder(t *testing.T) {
	identity := "0x000000000000000000000000000000000000000b"
	params := httprouter.Params{{Key: "iden", Value: identity}}
	reqBody := contract.OrderRequest{
		PriceAmount:     1,
		PriceCurrency:   "BTC",
		ReceiveCurrency: "BTC",
	}

	mock := &mockPilvytis{
		identity: identity,
		resp:     newMockPilvytisResp(1, identity, "BTC", "BTC", 1),
	}
	handler := NewPilvytisEndpoint(mock).CreatePaymentOrder

	mb, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(
		http.MethodPost,
		"/test",
		bytes.NewBuffer(mb),
	)
	assert.NoError(t, err)

	handler(resp, req, params)
	assert.Equal(t, 200, resp.Code)
	assert.JSONEq(t,
		`{
   "id":1,
   "status":"pending",
   "identity":"0x000000000000000000000000000000000000000b",
   "price_amount":1,
   "price_currency":"BTC",
   "pay_amount":1,
   "pay_currency":"test",
   "payment_address":"0x00",
   "receive_amount":1,
   "receive_currency":"BTC"
}`,
		resp.Body.String(),
	)

}

func TestGetPaymentOrder(t *testing.T) {
	identity := "0x000000000000000000000000000000000000000b"
	id := 11
	params := httprouter.Params{
		{Key: "iden", Value: identity},
		{Key: "id", Value: fmt.Sprint(id)},
	}

	mock := &mockPilvytis{
		identity: identity,
		resp:     newMockPilvytisResp(id, identity, "BTC", "BTC", 1),
	}
	handler := NewPilvytisEndpoint(mock).GetPaymentOrder

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)
	assert.NoError(t, err)

	handler(resp, req, params)
	assert.Equal(t, 200, resp.Code)
	assert.JSONEq(t,
		fmt.Sprintf(`{
   "id":%d,
   "status":"pending",
   "identity":"0x000000000000000000000000000000000000000b",
   "price_amount":1,
   "price_currency":"BTC",
   "pay_amount":1,
   "pay_currency":"test",
   "payment_address":"0x00",
   "receive_amount":1,
   "receive_currency":"BTC"
}`, id),
		resp.Body.String(),
	)

}

func TestGetPaymentOrders(t *testing.T) {
	identity := "0x000000000000000000000000000000000000000b"
	params := httprouter.Params{
		{Key: "iden", Value: identity},
	}

	mock := &mockPilvytis{
		identity: identity,
		resp:     newMockPilvytisResp(1, identity, "BTC", "BTC", 1),
	}
	handler := NewPilvytisEndpoint(mock).GetPaymentOrders

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)
	assert.NoError(t, err)

	handler(resp, req, params)
	assert.Equal(t, 200, resp.Code)
	assert.JSONEq(t,
		`[{
   "id":1,
   "status":"pending",
   "identity":"0x000000000000000000000000000000000000000b",
   "price_amount":1,
   "price_currency":"BTC",
   "pay_amount":1,
   "pay_currency":"test",
   "payment_address":"0x00",
   "receive_amount":1,
   "receive_currency":"BTC"
}]`,
		resp.Body.String(),
	)

}
