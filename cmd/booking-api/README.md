# Booking API projector

The Booking API projector is the component that relies on consuming from different topics and recreating the state of the entities inherent to the topic. Currently the Booking API projector projects these entities:

 - Booking (source: booking topic)
 - Booking References ( source: booking reference topic from one of the components that reads a lowribeck file and transforms it into booking references )
 - Occupancies ( source: energy-platform )
 - Service State ( source: energy-platform )
 - Site (source: energy-platform )

# Booking Big Query Indexer

The booking big query indexer consumes events from the booking topic and indexes them in big query. Depending on the nature of the event it will populate to its corresponding table.

# Booking API server

The booking API server is the component that is responsible to establish the communication with its Frontend counterpart to fulfill the smart-booking journey. The booking API server is responsible to retrieve user's information, user's bookings, user's address
and relay the communication for booking requests to lowribeck and interpret the response.

The API is composed by various requests:
 - Get Customer Contact Details
 - Get Customer Site Address
 - Get Customer Bookings
 - Get Available Slots
 - Create Booking
 - Reschedule Booking

The Booking API gRPC server can return different types of error codes. These error codes are also supplied with an error message to give more context to the nature of the error.
The nature of these errors can be:

 - User-generated errors
	 - In this case, the error is a result from the client that requests this server.
	 - It can be a missing parameter such as an absent request field (AccountID, Date Requested At)
 - Internal error
	 - Internally, the booking API has two different repositories for data: the eligibility API and the local projection which consists in a database, in case these components fail, the booking API server will return a generic Internal error.
 - Lowri-Beck wrapper error
	 - The Lowri-Beck wrapper error means that something has failed when doing a request to Lowri-Beck, there are different types of errors: bad parameter supplied, internal error, duplication error, not existent error.

The Booking API gRPC must handle these errors to the client applications in a way that all the inherent logic is abstracted and these applications know how to handle failures. (Retry with different parameters, Impossibility to continue, Try Again Later)

## Composite Types

### Booking Slot

| Field | Type/Description |
| -- | -- |
| Date | A string containing a date in "yyyy-mm-dd" format |
| Start Time | An int representing an hour in the day |
| End Time | An int representing an hour in the day |

### Vulnerabilities

| Field | Type/Description |
| -- | -- |
| Vulnerabilities | A list of Vulnerability enums |
| Other | A freeform text field |

### ContactDetails

## Get Customer Contact Details
The Get Customer Contact Details will return the supplied account ID's contact details:

### Request
| Field | Type |
| -- | -- |
| AccountID | String |

### Response
| Field | Type |
| -- | -- |
| Title | String |
| First Name | String |
| Last Name | String |
| Mobile Number | String |
| Email | String |

This request relies on the account-platform's gRPC server to provide this information.



### Error Codes & Description
|gRPC Error Code  | Description  |
|--|--|
| Internal | When there is an internal error it means something wrong happened during the database access or the account-platform gRPC server call. |
| InvalidArgument | The provided account ID is either missing or invalid(empty). |
| NotFound | This error can occur when the request done to the account-platform gRPC server results in an empty response.  |


## Get Customer Site Address
The Get Customer Site Address will return the supplied account ID's site address for the booking they are attempting to schedule(or reschedule).

### Error Codes & Description
|gRPC Error Code  | Description  |
|--|--|
| Internal | When there is an internal error it means something wrong happened during the database access or the gRPC eligibility server call. |
| InvalidArgument | The provided account ID is either missing or invalid(empty). |
| NotFound | This error can occur when the user does not have any eligible site address for a smart meter installation(none of his occupancies are eligible for a smart booking appointment) however this error is highly unlikely to happen due to the smart-booking journey only allowing to be started once the account has at least one eligible occupancy. However, in a stateless perspective this should be handled.  |

## Get Customer Bookings
The Get Customer Bookings calls the internal database projection to retrieve the user's bookings.
It takes in the account ID and the result should be a list of bookings.

### Error Codes & Description 
|gRPC Error Code  | Description  |
|--|--|
| Internal | When there is an internal error it means something wrong happened during the database access. |
| InvalidArgument | The provided account ID is either missing or invalid(empty). |

## Get Available Slots
The Get Available Slots is a call to Lowri-Beck that takes the following request parameters:

| Field | Type/Description |
| -- | -- |
| AccountID | String containing user's account ID |
| From | String containing a date in a yyyy-mm-dd format |
| To | String containing a date in a yyyy-mm-dd format  |

The response parameters will be a list of Booking Slots.

More information can be found [here](https://github.com/utilitywarehouse/energy-contracts/blob/master/protos/smart_booking/booking/v1/api.proto).

### Error Codes & Description 
|gRPC Error Code  | Description  |
|--|--|
| Internal | When an internal error occurs, this can be interpreted in many ways. The nature of this failure can derive from a problem with the database query, a gRPC call to the eligibility API or a failure to query the Lowri-Beck wrapper. More information can be found in the error message. |
| NotFound | This error can only come from the lowri-beck wrapper and it means the request made to Lowri-Beck resulted in a NotFound. Due to the nature of the call it can be interpreted as: no booking slots were found for the provided parameters. |
| OutOfRange | An out of range error means that for the supplied From and To dates in the request, no booking slots were possible to be found. |
| InvalidArgument | Any of the previously mentioned fields in the request is missing or has an empty value. |

## Create Booking
The Create Booking results in a call to Lowri-Beck being made and a booking being created. The Create Booking takes in the following parameters:

| Field | Type/Description |
| -- | -- |
| AccountID | String containing a user's account ID |
| Booking Slot | [BookingSlot](#composite-types) |
| Vulnerabilities | [Vulnerabilities](#composite-types) |
| Contact Details | [ContactDetails](#composite-types) |
| Platform | the platform that is creating the request, can be mobile, web, my-app. It is not relevant for lowri-beck but it is important internally to understand where the request is being made from. |
 
 The response parameter will be the internally generated booking ID (The Booking API generated uuid) in case of success.
### Error Codes & Description 
|gRPC Error Code  | Description  |
|--|--|
| Internal | The nature of this failure can derive from a problem with the database query, a gRPC call to the eligibility API or a failure to query the Lowri-Beck wrapper. More information can be found in the error message. |
| NotFound | This error can only come from the lowri-beck wrapper and it means the request made to Lowri-Beck resulted in a NotFound. |
| OutOfRange | An out of range error means that for the supplied Booking Slot date in the request, it was impossible to create a booking. It might be due to the attempted create booking date is too soon and Lowri-Beck does not want to handle. |
| InvalidArgument | Any of the previously mentioned fields in the request is missing or has an empty value.  It can also mean the supplied Booking slot's date OR time is not valid (from LowriBeck's perspective). |
| AlreadyExists | This error should only occur when the request to create a booking conflicts with an actual booking already existing in LowriBeck's end. Can be interpreted as "booking duplicated". |

## Reschedule Booking
The Reschedule Booking results in a call to Lowri-Beck being made and a previously created booking having its booking slot changed. The Reschedule Booking takes in the following parameters:

### Request

| Field | Type/Description |
| -- | -- |
| AccountID | A string containing the account ID of the user |
| Booking Slot | [BookingSlot](#composite-types) |
 
 The response parameter will be the internal booking ID (The Booking API generated uuid during a Create Booking call) in case of success.
 
### Error Codes & Description 
The error codes for Reschedule Booking are very similar to Create Booking, since the logic for a creation and a rescheduling from LowriBeck's wrapper is the same.
