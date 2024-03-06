package service

import (
	"log"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
)

type classicService struct {
	client *server.APIClient
}

func NewClassicService(apiKey *ncloud.APIKey) classicService {
	return classicService{client: server.NewAPIClient(server.NewConfiguration(apiKey))}
}

func (s classicService) GetServerImageProductList() []*vserver.Product {
	req := server.GetServerImageProductListRequest{}

	if r, err := s.client.V2Api.GetServerImageProductList(&req); err != nil {
		log.Println("err")
		log.Println(err)
	} else {
		return convertVpcProducts(r.ProductList)

	}
	return nil
}

func (s classicService) GetServerProductList(p *vserver.Product) []*vserver.Product {
	req := server.GetServerProductListRequest{
		ServerImageProductCode: p.ProductCode,
	}

	if r, err := s.client.V2Api.GetServerProductList(&req); err != nil {
		log.Println(err)
	} else {
		return convertVpcProducts(r.ProductList)
	}

	return nil
}

func convertVpcProducts(products []*server.Product) []*vserver.Product {
	m := make([]*vserver.Product, 0)
	for _, r := range products {
		m = append(m, &vserver.Product{
			ProductCode:          r.ProductCode,
			ProductName:          r.ProductName,
			ProductDescription:   r.ProductDescription,
			ProductType:          &vserver.CommonCode{Code: r.ProductType.Code, CodeName: r.ProductType.CodeName},
			BaseBlockStorageSize: r.BaseBlockStorageSize,
		})
	}

	return m
}
