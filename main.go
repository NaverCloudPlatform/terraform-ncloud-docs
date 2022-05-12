package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sync"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/NaverCloudPlatform/terraform-ncloud-docs/service"
	"gopkg.in/yaml.v3"
)

const GIGA_BYTE = 1024 * 1024 * 1024

type Accounts struct {
	Accounts []struct {
		Domain    string `yaml:"domain"`
		Region    string `yaml:"region"`
		AccessKey string `yaml:"accessKey"`
		SecretKey string `yaml:"secretKey"`
		ApiUrl    string `yaml:"apiUrl"`
	} `yaml:"accounts"`
}

func main() {
	filename, _ := filepath.Abs("account.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	var accounts Accounts
	err = yaml.Unmarshal(yamlFile, &accounts)
	if err != nil {
		return
	}

	var wg sync.WaitGroup

	// PUB, FIN, GOV 통합을 위한 새로운 구조체 변수들
	vpcImageProductsV2 := []*ProductV2{}
	vpcServerProductsV2 := map[string]*ServerProductsV2{}
	classicImageProductsV2 := []*ProductV2{}
	classicServerProductsV2 := map[string]*ServerProductsV2{}

	// PUB, FIN, GOV 순회하여 실행
	for _, account := range accounts.Accounts {
		apiKeys := &ncloud.APIKey{
			AccessKey: account.AccessKey,
			SecretKey: account.SecretKey,
		}
		os.Setenv("NCLOUD_API_GW", account.ApiUrl)

		// VPC 이미지상품 데이터 생성
		vpcService := service.NewVpcService(apiKeys)
		vpcImageProducts := vpcService.GetServerImageProductList(account.Region)
		vpcImageProductsV2 = UpdateProductsWithDomain(vpcImageProductsV2, vpcImageProducts, account.Domain)

		// VPC 서버상품 데이터 생성
		for _, r := range vpcImageProducts {
			wg.Add(1)
			go func(r *vserver.Product) {
				defer wg.Done()
				vpcServerProducts := vpcService.GetServerProductList(r, account.Region)

				if _, isExist := vpcServerProductsV2[*r.ProductCode]; isExist {
					vpcServerProductsV2[*r.ProductCode].productsV2 = UpdateProductsWithDomain(vpcServerProductsV2[*r.ProductCode].productsV2, vpcServerProducts, account.Domain)
				} else {
					productsV2 := []*ProductV2{}
					vpcServerProductsV2[*r.ProductCode] = &ServerProductsV2{
						productsV2:              UpdateProductsWithDomain(productsV2, vpcServerProducts, account.Domain),
						imageProductName:        *r.ProductName,
						imageProductDescription: *r.ProductDescription,
						imageProductCode:        *r.ProductCode,
					}
				}
			}(r)
		}
		wg.Wait()

		// FIN 인 경우 classic 상품을 조회하지 않음
		if account.Domain != "Fin" {
			// 클래식 이미지상품 데이터 생성
			classicService := service.NewClassicService(apiKeys)
			classicImageProducts := classicService.GetServerImageProductList()
			classicImageProductsV2 = UpdateProductsWithDomain(classicImageProductsV2, classicImageProducts, account.Domain)

			// 클래식 서버상품 데이터 생성
			for _, r := range classicImageProducts {
				wg.Add(1)
				go func(r *vserver.Product) {
					defer wg.Done()
					classicServerProducts := classicService.GetServerProductList(r)

					if _, isExist := classicServerProductsV2[*r.ProductCode]; isExist {
						classicServerProductsV2[*r.ProductCode].productsV2 = UpdateProductsWithDomain(classicServerProductsV2[*r.ProductCode].productsV2, classicServerProducts, account.Domain)
					} else {
						productsV2 := []*ProductV2{}
						classicServerProductsV2[*r.ProductCode] = &ServerProductsV2{
							productsV2:              UpdateProductsWithDomain(productsV2, classicServerProducts, account.Domain),
							imageProductName:        *r.ProductName,
							imageProductDescription: *r.ProductDescription,
							imageProductCode:        *r.ProductCode,
						}
					}
				}(r)
			}
			wg.Wait()
		}
	}

	// product name 기준으로 이미지상품 정렬
	sort.SliceStable(vpcImageProductsV2, func(i, j int) bool {
		return strings.ToUpper(*vpcImageProductsV2[i].ProductName) < strings.ToUpper(*vpcImageProductsV2[j].ProductName)
	})

	sort.SliceStable(classicImageProductsV2, func(i, j int) bool {
		return strings.ToUpper(*classicImageProductsV2[i].ProductName) < strings.ToUpper(*classicImageProductsV2[j].ProductName)
	})

	// product name 기준으로 서버상품 정렬
	for _, serverProducts := range vpcServerProductsV2 {
		sort.SliceStable(serverProducts.productsV2, func(i, j int) bool {
			return strings.ToUpper(*serverProducts.productsV2[i].ProductName) < strings.ToUpper(*serverProducts.productsV2[j].ProductName)
		})
	}

	for _, serverProducts := range classicServerProductsV2 {
		sort.SliceStable(serverProducts.productsV2, func(i, j int) bool {
			return strings.ToUpper(*serverProducts.productsV2[i].ProductName) < strings.ToUpper(*serverProducts.productsV2[j].ProductName)
		})
	}

	// 이미지상품 md 파일 작성
	var b bytes.Buffer
	b.WriteString(createMarkdownImages(vpcImageProductsV2, "Server Image Products (VPC)", "vpc_products"))
	b.WriteString("\n")
	b.WriteString(createMarkdownImages(classicImageProductsV2, "Server Image Products (Classic)", "classic_products"))

	createFile(b.String(), "docs/server_image_product.md")

	// 서버상품 md 파일 작성
	for _, serverProductsV2 := range vpcServerProductsV2 {
		wg.Add(1)
		go func(serverProductsV2 *ServerProductsV2) {
			defer wg.Done()
			md := createMarkdownProducts(fmt.Sprintf("Server products of image(%s) - %s", serverProductsV2.imageProductDescription, serverProductsV2.imageProductCode), serverProductsV2.productsV2)
			replacer := strings.NewReplacer(" ", "+", "/", "")
			createFile(md, "docs/vpc_products/"+replacer.Replace(serverProductsV2.imageProductName)+".md")
		}(serverProductsV2)
	}
	wg.Wait()

	for _, serverProductsV2 := range classicServerProductsV2 {
		wg.Add(1)
		go func(serverProductsV2 *ServerProductsV2) {
			defer wg.Done()
			md := createMarkdownProducts(fmt.Sprintf("Server products of image(%s) - %s", serverProductsV2.imageProductDescription, serverProductsV2.imageProductCode), serverProductsV2.productsV2)
			replacer := strings.NewReplacer(" ", "+", "/", "")
			createFile(md, "docs/classic_products/"+replacer.Replace(serverProductsV2.imageProductName)+".md")
		}(serverProductsV2)
	}
	wg.Wait()

}

func createMarkdownImages(productsV2 []*ProductV2, title string, productPath string) string {
	var b bytes.Buffer
	b.WriteString("### " + title + "\n\n")
	b.WriteString("Description | Image code | Type | B/S Size(GB) | Pub | Fin | Gov |\n")
	b.WriteString("-- | -- | -- | -- | -- | -- | -- |\n")

	for _, r := range productsV2 {
		b.WriteString(fmt.Sprintf("[%s](%s/%s.md) | %s | %s | %d | %s | %s | %s |\n", ncloud.StringValue(r.ProductDescription), productPath, url.QueryEscape(strings.Replace(ncloud.StringValue(r.ProductName), "/", "", -1)), ncloud.StringValue(r.ProductCode), ncloud.StringValue(r.ProductType.CodeName), ncloud.Int64Value(r.BaseBlockStorageSize)/GIGA_BYTE, r.Pub, r.Fin, r.Gov))
	}

	return b.String()
}

func createMarkdownProducts(title string, productsV2 []*ProductV2) string {
	var b bytes.Buffer
	b.WriteString("### " + title + "\n\n")
	b.WriteString("Description | Product code | Type | Pub | Fin | Gov |\n")
	b.WriteString("-- | -- | -- | -- | -- | -- |\n")

	for _, r := range productsV2 {
		b.WriteString(fmt.Sprintf("%s | %s | %s | %s | %s | %s |\n", ncloud.StringValue(r.ProductDescription), ncloud.StringValue(r.ProductCode), ncloud.StringValue(r.ProductType.Code), r.Pub, r.Fin, r.Gov))
	}

	return b.String()
}

func createFile(s, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	n, err := file.Write([]byte(s))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(filename, "-", n, "bytes")
}

// 도메인 정보가 포함된 새로운 product 구조체
type ProductV2 struct {
	vserver.Product
	Pub string
	Fin string
	Gov string
}

// server product 배열에 image 속성을 더한 server product 리스트 구조체
type ServerProductsV2 struct {
	productsV2              []*ProductV2
	imageProductName        string
	imageProductDescription string
	imageProductCode        string
}

// product v2 배열에 product code 를 기준으로 새로운 product code이면 새로운 객체를, 기존의 product code이면 도메인만 업데이트
func UpdateProductsWithDomain(productsV2 []*ProductV2, products []*vserver.Product, domain string) []*ProductV2 {

	for _, product := range products {
		isExist := false
		for _, productV2 := range productsV2 {
			if *productV2.ProductCode == *product.ProductCode {
				isExist = true

				switch domain {
				case "Pub":
					productV2.Pub = "O"
				case "Fin":
					productV2.Fin = "O"
				case "Gov":
					productV2.Gov = "O"
				}
			}
		}
		if isExist == false {
			productV2 := &ProductV2{
				Product: *product,
				Pub:     "X",
				Fin:     "X",
				Gov:     "X",
			}
			switch domain {
			case "Pub":
				productV2.Pub = "O"
			case "Fin":
				productV2.Fin = "O"
			case "Gov":
				productV2.Gov = "O"
			}
			productsV2 = append(productsV2, productV2)
		}
	}
	return productsV2
}
