package industry

import (
	"context"
	"errors"
	"fmt"

	ecoesv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v1"
	xoservev1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/xoserve/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EcoesAPI interface {
	GetTechnicalDetailsByMPAN(ctx context.Context, req *ecoesv1.SearchByMPANRequest, opts ...grpc.CallOption) (*ecoesv1.TechnicalDetailsResponse, error)
	GetRelAddresses(ctx context.Context, in *ecoesv1.SearchByMPANRequest, opts ...grpc.CallOption) (*ecoesv1.GetRELAddressesResponse, error)
	GetRelatedMPANs(ctx context.Context, in *ecoesv1.SearchByMPANRequest, opts ...grpc.CallOption) (*ecoesv1.GetRelatedMPANsResponse, error)
}

type XoserveAPI interface {
	GetSwitchDataByMPRN(ctx context.Context, req *xoservev1.SearchByMPRNRequest, opts ...grpc.CallOption) (*xoservev1.TechnicalDetailsResponse, error)
}

func New(ecoesAPI EcoesAPI, xoserveAPI XoserveAPI, machine auth.MachineInjector) api.Client {
	return &IndustryClients{
		ecoesAPI:   ecoesAPI,
		xoserveAPI: xoserveAPI,
		machine:    machine,
	}
}

type IndustryClients struct {
	ecoesAPI   EcoesAPI
	xoserveAPI XoserveAPI

	machine auth.MachineInjector
}

var (
	ErrMultipleResults = errors.New("multiple results")
)

func (c *IndustryClients) getMeterpointByMPAN(ctx context.Context, mpan string) (*ecoesv1.TechnicalDetailsResponse, error) {
	resp, err := c.ecoesAPI.GetTechnicalDetailsByMPAN(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: mpan,
	})

	if err != nil {
		switch status.Code(err) {
		default:
			return nil, fmt.Errorf("unknown error: %s", err.Error())
		case codes.NotFound:
			return nil, api.ErrNotFound
		}
	}

	return resp, nil
}

func (c *IndustryClients) GetMeterpointByMPAN(ctx context.Context, mpan string) (*sitev1.ElectricityMeterpointDetails, *addressv1.Address, error) {
	ctx, span := tracing.Tracer().Start(c.machine.ToCtx(ctx), "industry.get_meterpoint_by_mpan")
	defer span.End()

	var meterpoint *sitev1.ElectricityMeterpointDetails

	m, err := c.getMeterpointByMPAN(ctx, mpan)
	if err != nil {
		return nil, nil, tracing.RecordAndReturnError(span, err)
	}

	meterpoint, err = mapElectricityMeterpointDetails(m)
	if err != nil {
		return nil, nil, tracing.RecordAndReturnError(span, err)
	}

	rel, err := c.getRelAddresses(ctx, mpan)
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		return nil, nil, tracing.RecordAndReturnError(span, err)
	}

	return meterpoint, mapElectricitySite(rel), nil
}

func (c *IndustryClients) GetMeterpointByMPRN(ctx context.Context, mprn string) (*sitev1.GasMeterpoint, *addressv1.Address, error) {
	ctx, span := tracing.Tracer().Start(c.machine.ToCtx(ctx), "industry.get_meterpoint_by_mprn")
	defer span.End()

	resp, err := c.getMeterpointByMPRN(ctx, mprn)
	if err != nil {
		return nil, nil, tracing.RecordAndReturnError(span, err)
	}

	meterpoint, err := mapGasMeterpointDetails(resp)
	if err != nil {
		return nil, nil, tracing.RecordAndReturnError(span, err)
	}

	return meterpoint, resp.GetAddress(), nil
}
