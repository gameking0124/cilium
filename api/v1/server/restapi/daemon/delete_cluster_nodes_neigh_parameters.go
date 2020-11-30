// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2017-2020 Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package daemon

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/cilium/cilium/api/v1/models"
)

// NewDeleteClusterNodesNeighParams creates a new DeleteClusterNodesNeighParams object
// no default values defined in spec.
func NewDeleteClusterNodesNeighParams() DeleteClusterNodesNeighParams {

	return DeleteClusterNodesNeighParams{}
}

// DeleteClusterNodesNeighParams contains all the bound params for the delete cluster nodes neigh operation
// typically these are obtained from a http.Request
//
// swagger:parameters DeleteClusterNodesNeigh
type DeleteClusterNodesNeighParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: body
	*/
	Request *models.NodeNeighRequest
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewDeleteClusterNodesNeighParams() beforehand.
func (o *DeleteClusterNodesNeighParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.NodeNeighRequest
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("request", "body", ""))
			} else {
				res = append(res, errors.NewParseError("request", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Request = &body
			}
		}
	} else {
		res = append(res, errors.Required("request", "body", ""))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}