# Duffel API Go Client

A Go (golang) client library for the [Duffel](https://duffel.com) API implemented by the Airheart team.

[![Tests](https://github.com/airheartdev/duffel/actions/workflows/ci.yaml/badge.svg)](https://github.com/airheartdev/duffel/actions/workflows/ci.yaml)

## Installation

**Requires at least Go 1.18-rc1 since we use generics on the internal API client**

```shell
go get github.com/airheartdev/duffel
```

## Usage examples

See the [examples/\*](/examples/) directory

## Implementation status:

To maintain simplicity and ease of use, this client library is hand-coded (instead of using Postman to Go code generation) and contributions are greatly apprecicated.

- [x] Most API types
- [x] API Client
- [x] Error handling
- [x] Pagination _(using iterators)_
- [x] Rate Limiting _(automatically throttles requests to stay under limit)_
- [x] Offer Requests
  - [x] Create offer request and return offer
  - [x] Get offer by ID
  - [x] List all offers
- [x] Offers
- [ ] Orders
- [ ] Payments
- [ ] Seat Maps
- [ ] Order Cancellations
- [ ] Order Change Requests
- [ ] Order Change Offers
- [ ] Order Changes

## License

MIT.
