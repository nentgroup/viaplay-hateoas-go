# API Guidelines

## 1 Introduction

This document describes the HATEOAS format that is used in the MTG API.

## 2 Requirements

The keywords "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC2119].

## 3 MTG Json Documents

A MTG JSON Document uses the json-format described in [RFC4627] and has the media type "application/mtg+json".

Its root object MUST be a Resource Object.

For example:

```json
GET /orders/523 HTTP/1.1
Host: example.org
Accept: application/mtg+json

HTTP/1.1 200 OK
Content-Type: application/mtg+json

{
    "links": {
        "self": { 
            "href": "/orders/523",
            "method": "GET"
        },
        "warehouse": { 
            "href": "/warehouse/56", 
            "method": "GET",
            "returnType": "warehouse"
        },
    },
    "data": {
        "name": "hat",
        "type": {
            "size": "M",
            "colour": "black"
        },
        "currency": "USD"
    },
    "embedded": {
        "all-orders": [
            { 
                "href": "/orders/324",
                "method": "GET"
            },
            { 
                "href": "/orders/523",
                "method": "GET"
            }
        ]
    },
}
```

The above example represents an order resource with the URI "/orders/523". It has links to warehouse and invoice and its own state inside the data property.

## 4 Resource Objects

A resource object MUST NOT have any other properties than "links", "data", "embedded" and "error".

(i). "links" contains named links to other resources. It is an object whose property names are link relation types (as defined by [RFC5988]) and values MUST be a Link Object (see below).

(ii). "data" contains all fields that are owned by this resource. The internal structure of this object has no constraints but lists should normally go into embedded instead, since each list item should be considered a separate resource (and therefore not owned by the resource representing the list itself).

(iii). "embedded" contains resources that are conceptually other resources but that are returned within this response as a convenience.

(iv). "error" MUST be populated if the HTTP status code is not 2XX.


## 5 Link Objects

"links" is a flat map of links where each object is a Link Object that represents a hyperlink from the containing resource to a URI. A link shall not be an array, so if there is a need to expose a dynamic list of links, it should be put as a property under embedded. The link can be both external or a link to another object within the MTG API. Link Object have the following properties:

### 5.1 href (REQUIRED)

Its value is either a URI [RFC3986](https://www.rfc-editor.org/rfc/rfc3986) or a URI Template [RFC6570](https://www.rfc-editor.org/rfc/rfc6570).

If the URI refers to another MSP resource the URI SHOULD be relative (ie start with /), otherwise it SHOULD be absolute.

### 5.2 method (OPTIONAL)

This property tells the client which http method to use for when calling the endpoint specified in the href-property. If not specified GET SHOULD be used.

Its value SHOULD be one of one of "GET", "POST", "PUT", "PATCH" or "DELETE".

### 5.3 title (OPTIONAL)

Optional localized title that MAY be shown to the end user.

### 5.4 returnType (REQUIRED IF LINK IS RELATIVE (MSP CALLS))

The name of the object (protobuf type) that will be returned.

### 5.5 requestBodyType (OPTIONAL)

The name of the object (protobuf type) that SHOULD be provided if the link is to a MSP service and method is POST or PUT and it expects a post-body.

### 5.6 deprecated (OPTIONAL)

Its presence indicates that the link is to be removed at a future date. Its value SHOULD provide further information about the deprecation.

## 6 Embedded

Other resources that normally could have been retrieved by following links, but that are included in this response for convenience. Each resource in embedded should have the same format as other resources, ie the resource object MUST NOT have any other properties than "links", "data", "embedded" and "error".


## 7 Error Objects

TBD

## 8 FAQ

**Q**: Can I nest objects under "links"?  
**A**: No. Links is a flat map or Link Objects

**Q**: Can I add more properties on a Link Object?  
**A**: No only the specified properties are allowed

**Q**: I have an endpoint that returns a list, should the objects go into "data" or "embedded"?  
**A**: They should go into embedded since each object can have it's own link etc. In this use-case "data" typically only contains the number of items and similar meta data. That is the only data "owned" by the list endpoint.


**Q**: I need to return a list of links but "links" is a flat map.  
**A**: Use embedded instead. This also enables additional data to be added per list item later if needed.

## 9 Authors
This specification was created by [Morten Diesen](morten.diesen@viaplay.com), [Joakim Rapp](joakim.rapp@viaplay.com), [Bernhard Wang](bernhard.hettman@viaplay.com) and [John Jedborn](john.jedborn@viaplay.com) at MTG AB, Sweden, 2016.