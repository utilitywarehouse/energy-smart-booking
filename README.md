# energy-smart-booking

This repo is a collection of services and shared code related to the smart
meter installation booking journey.

## Repo structure
**cmd/** - all individual commands, implementations, and business logic.

**internal/** - common utilities like database connection code, shared domain, types, repositories, shared test utilities, etc

**Makefile** - single makefile taking in source_files and services as arguments to build a specific executable
    example: make SERVICE=energy-smart-booking-opt-out SOURCE_FILES=cmd/opt-out/ energy-smart-booking-opt-out

## GitHub workflows

1. Reusable build service [build_service_reusable_workflow](https://github.com/utilitywarehouse/energy-smart-booking/blob/master/.github/workflows/build_service_reusable_workflow.yml)
    Run install, lint, test and docker build for a given service.
    This workflow is intended to run only as a callable action by a specific booking service workflow or to build all in tandem.

2. Callable build-smart-booking workflow [smart_booking_workflow](https://github.com/utilitywarehouse/energy-smart-booking/blob/master/.github/workflows/energy_smart_booking.yml)
    Calls reusable build service for all booking service executables.
    Triggered when there is any change in the project other than just cmd/$BOOKING_SERVICE.

3. CI entrypoint (https://github.com/utilitywarehouse/energy-smart-booking/blob/master/.github/workflows/ci.yml).
    Checks the files changed compared to previous commit to determine what build it needs to run.
    Logic:
    - use tj-actions/changed-files@v35 to get all changed files
    - go through each file and:
      - if file belongs to cmd/$BOOKING_SERVICE_NAME - set job output as needs.check.outputs.'$BOOKING_SERVICE_NAME' = true
      - if file doesn't belong to any cmd/** - set job output as needs.check.outputs.booking_services = true

    - job entry for each specific fabricator checking if it should run using:

    `needs.check.outputs.booking_services != 'true' && needs.check.outputs.opt_out == 'true'`

    - job entry to build all booking checking if it should run using:

    `needs.check.outputs.booking_services == 'true'`

    By using **workflow_dispatch** we avoid duplication of workflows triggered when there is a PR opened and commits are done on the feature implementation.
   - every branch is built on push without a need to create a PR
   - when you create PR this “branch build” is already picked up by checks
   - when you update PR it is built, because branch has changed
   - master is built on branch merge after PR

## Testing locally

Install gomock:
```
    go install github.com/golang/mock/mockgen@v1.6.0
```

Generate the mocks:
```
    go generate ./...
```

## Booking Services Overview

### LowriBeck API Wrapper
Calls the third party LowriBeck API for booking smart meter installation appointments, for more information see the [API specification](https://wiki.uw.systems/posts/industry-ap-is-wip-lmd7g5jx#hdgcj-lowri-beck-api). There are two endpoints, the first is GetAvailableSlots which returns all the available slots for a postcode and the following GRPC statuses:
| HTTP | gRPC | Description |
| --- | --- | --- |
| 200 | OK | No error. |
| 400 | INVALID_ARGUMENT | Client specified an invalid argument. Check error message for more information. |
| 400 | OUT_OF_RANGE | Booking request sent outside agreed time parameter. (As we don't send any dates to LB, it's not clear if this error will ever occur) |
| 404 | NOT_FOUND | No available slots for requested postcode. |
| 500 | INTERNAL | Internal server error. Typically a server bug. |


The second endpoint is CreateBooking which returns whether a booking was successful and the following GRPC statuses:
| HTTP | gRPC | Description |
| --- | --- | --- |
| 200 | OK | No error. |
| 400 | INVALID_ARGUMENT | Client specified an invalid argument. Check error message for more information will be either postcode, reference, site, appointment date or appointment time. |
| 400 | OUT_OF_RANGE | Booking request sent outside agreed time parameter. |
| 404 | NOT_FOUND | No available slots for requested postcode. |
| 409 | ALREADY_EXISTS | Duplicate booking exists. |
| 500 | INTERNAL | Internal server error. Typically a server bug. |


## Booking API

Please refer to this(https://github.com/utilitywarehouse/energy-smart-booking/blob/master/cmd/booking-api/README.md) README.


### opt-out
Handles opt-outs for smart meter installation. 
The purpose of the service is to be used in-house (no customer facing exposure) to mark the 
customers that don't wish to be campaigned about or go through a smart meter installment process. 
The UI for this service can be found at https://energy-smart-booking-opt-out-ui.prod.aws.uw.systems/
Events AccountBookingOptOutAdded/RemovedEvent are published every time we update the list of opt-outs,
either by adding or removing an account from there. 


### click-generator
Click generator is a service used for testing smart booking journey when we use pre authenticated
links. 
The solution uses https://github.com/utilitywarehouse/click.uw.co.uk for link generation and exposes 
and endpoint inside UW via ingress definition:

DEV: https://smart-booking-click-api.dev.merit.uw.systems/generate?type=auth

PROD: https://smart-booking-click-api.prod.aws.uw.systems/generate?type=auth

Example of usage:
POST https://smart-booking-click-api.dev.merit.uw.systems/generate?type=auth

Body: { "account_number": "7821689" }
