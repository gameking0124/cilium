// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2017-2020 Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package daemon

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PutClusterNodesNeighHandlerFunc turns a function with the right signature into a put cluster nodes neigh handler
type PutClusterNodesNeighHandlerFunc func(PutClusterNodesNeighParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PutClusterNodesNeighHandlerFunc) Handle(params PutClusterNodesNeighParams) middleware.Responder {
	return fn(params)
}

// PutClusterNodesNeighHandler interface for that can handle valid put cluster nodes neigh params
type PutClusterNodesNeighHandler interface {
	Handle(PutClusterNodesNeighParams) middleware.Responder
}

// NewPutClusterNodesNeigh creates a new http.Handler for the put cluster nodes neigh operation
func NewPutClusterNodesNeigh(ctx *middleware.Context, handler PutClusterNodesNeighHandler) *PutClusterNodesNeigh {
	return &PutClusterNodesNeigh{Context: ctx, Handler: handler}
}

/*PutClusterNodesNeigh swagger:route PUT /cluster/nodes/neigh daemon putClusterNodesNeigh

Insert node as a neighbor into cluster

Inserts a node as a neighbor into the cluster, rather than a full node
(running Cilium). This operation will create a permanent entry in the
current node's neighbor table.


*/
type PutClusterNodesNeigh struct {
	Context *middleware.Context
	Handler PutClusterNodesNeighHandler
}

func (o *PutClusterNodesNeigh) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutClusterNodesNeighParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}