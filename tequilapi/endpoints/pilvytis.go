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

package endpoints

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/pilvytis"
	"github.com/mysteriumnetwork/node/tequilapi/contract"
	"github.com/mysteriumnetwork/node/tequilapi/utils"
	"github.com/pkg/errors"
)

type api interface {
	CreatePaymentOrder(i identity.Identity, priceAmount float64, priceCurr, recvCurr string, ln bool) (pilvytis.OrderResp, error)
	GetPaymentOrder(i identity.Identity, oid uint64) (pilvytis.OrderResp, error)
	GetPaymentOrders(i identity.Identity) ([]pilvytis.OrderResp, error)
}

type pilvytisEndpoint struct {
	api api
}

// NewPilvytisEndpoint returns pilvytis endpoints.
func NewPilvytisEndpoint(pil api) *pilvytisEndpoint {
	return &pilvytisEndpoint{
		api: pil,
	}
}

// CreatePaymentOrder creates a new payment order.
//
// swagger:operation POST /identity/{iden}/pilvytis/order Order createOrder
// ---
// summary: Create order
// description: Takes the given data and tries to create a new payment order in the pilvytis service.
// parameters:
// - name: iden
//   in: path
//   description: Identity for which to create an order
//   type: string
//   required: true
// - in: body
//   name: body
//   description: Required data to create a new order
//   schema:
//     $ref: "#/definitions/OrderRequest"
// responses:
//   200:
//     description: Order object
//     schema:
//       "$ref": "#/definitions/OrderResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorMessageDTO"
func (e *pilvytisEndpoint) CreatePaymentOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var req contract.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, errors.Wrap(err, "failed to parse order req"), http.StatusBadRequest)
		return
	}

	var identity identity.Identity
	identity.Address = params.ByName("iden")
	resp, err := e.api.CreatePaymentOrder(
		identity,
		req.PriceAmount,
		req.PriceCurrency,
		req.ReceiveCurrency,
		req.LightningNetwork)
	if err != nil {
		utils.SendError(w, err, http.StatusInternalServerError)
		return
	}

	utils.WriteAsJSON(contract.NewOrderResponse(resp), w)
}

// GetPaymentOrder returns a payment order which maches a given ID and identity.
//
// swagger:operation GET /identity/{iden}/pilvytis/order/{id} Order getOrder
// ---
// summary: Get order
// description: Gets an order for a given identity and order id combo from the pilvytis service
// parameters:
// - name: iden
//   in: path
//   description: Identity for which to get an order
//   type: string
//   required: true
// - name: id
//   in: path
//   description: Order id
//   type: integer
//   required: true
// responses:
//   200:
//     description: Order object
//     schema:
//       "$ref": "#/definitions/OrderResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorMessageDTO"
func (e *pilvytisEndpoint) GetPaymentOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	if id == "" {
		utils.SendError(w, errors.New("missing ID param"), http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		utils.SendError(w, errors.New("can't parse order ID as uint"), http.StatusBadRequest)
		return
	}

	var identity identity.Identity
	identity.Address = params.ByName("iden")

	resp, err := e.api.GetPaymentOrder(identity, orderID)
	if err != nil {
		utils.SendError(w, err, http.StatusInternalServerError)
		return
	}

	utils.WriteAsJSON(contract.NewOrderResponse(resp), w)
}

// GetPaymentOrder returns a payment order which maches a given ID and identity.
//
// swagger:operation GET /identity/{iden}/pilvytis/order Order getOrders
// ---
// summary: Get all orders for identity
// description: Gets all orders for a given identity from the pilvytis service
// parameters:
// - name: iden
//   in: path
//   description: Identity for which to get orders
//   type: string
//   required: true
// responses:
//   200:
//     description: Array of order objects
//     schema:
//       type: "array"
//       items:
//         "$ref": "#/definitions/OrderResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorMessageDTO"
func (e *pilvytisEndpoint) GetPaymentOrders(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var identity identity.Identity
	identity.Address = params.ByName("iden")

	resp, err := e.api.GetPaymentOrders(identity)
	if err != nil {
		utils.SendError(w, err, http.StatusInternalServerError)
		return
	}

	utils.WriteAsJSON(contract.NewOrdersResponse(resp), w)
}

// AddRoutesForPilvytis adds the pilvytis routers to the given router.
func AddRoutesForPilvytis(router *httprouter.Router, pilvytis api) {
	pil := NewPilvytisEndpoint(pilvytis)
	router.POST("/identity/:iden/pilvytis/order", pil.CreatePaymentOrder)
	router.GET("/identity/:iden/pilvytis/order/:id", pil.GetPaymentOrder)
	router.GET("/identity/:iden/pilvytis/order", pil.GetPaymentOrders)
}
