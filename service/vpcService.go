package service

import (
	"log"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
)

type vpcService struct {
	client *vserver.APIClient
}

func NewVpcService(apiKey *ncloud.APIKey) vpcService {
	return vpcService{client: vserver.NewAPIClient(vserver.NewConfiguration(apiKey))}
}

func (s vpcService) GetServerImageProductList() []*vserver.Product {
	req := vserver.GetServerImageProductListRequest{
		RegionCode: ncloud.String("KR"),
	}

	if r, err := s.client.V2Api.GetServerImageProductList(&req); err != nil {
		log.Println(err)
	} else {
		return r.ProductList

	}
	return nil
}

func (s vpcService) GetServerProductList(p *vserver.Product) []*vserver.Product {
	req := vserver.GetServerProductListRequest{
		RegionCode:             ncloud.String("KR"),
		ServerImageProductCode: ncloud.String(*p.ProductCode),
	}

	if r, err := s.client.V2Api.GetServerProductList(&req); err != nil {
		log.Println(err)
	} else {
		return r.ProductList
	}

	return nil
}
