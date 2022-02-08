package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/NaverCloudPlatform/terraform-ncloud-docs/service"
)

const GIGA_BYTE = 1024 * 1024 * 1024

func main() {
	apiKeys := ncloud.Keys()
	// apiKeys := &ncloud.APIKey{
	// 	AccessKey: "YOUR_ACCESS_KEY",
	// 	SecretKey: "YOUR_SECRET_KEY",
	// }
	vpcService := service.NewVpcService(apiKeys)
	classicService := service.NewClassicService(apiKeys)

	vpcImages := vpcService.GetServerImageProductList()
	classicImages := classicService.GetServerImageProductList()

	var b bytes.Buffer
	b.WriteString(createMarkdownImages(vpcImages, "Server Image Products (VPC)", "vpc_products"))
	b.WriteString("\n")
	b.WriteString(createMarkdownImages(classicImages, "Server Image Products (Classic)", "classic_products"))

	createFile(b.String(), "docs/server_image_product.md")

	var wg sync.WaitGroup
	for _, r := range vpcImages {
		wg.Add(1)
		go func(r *vserver.Product) {
			defer wg.Done()
			products := vpcService.GetServerProductList(r)
			md := createMarkdownProducts(fmt.Sprintf("Server products of image(%s) - %s", *r.ProductDescription, *r.ProductCode), products)
			createFile(md, "docs/vpc_products/"+*r.ProductName+".md")
		}(r)
	}

	for _, r := range classicImages {
		wg.Add(1)
		go func(r *vserver.Product) {
			defer wg.Done()
			products := classicService.GetServerProductList(r)
			md := createMarkdownProducts(fmt.Sprintf("Server products of image(%s) - %s", *r.ProductDescription, *r.ProductCode), products)
			createFile(md, "docs/classic_products/"+*r.ProductName+".md")
		}(r)
	}

	wg.Wait()
}

func createMarkdownImages(products []*vserver.Product, title string, productPath string) string {
	var b bytes.Buffer
	b.WriteString("### " + title + "\n\n")
	b.WriteString("Description | Image code | Type | B/S Size(GB)\n")
	b.WriteString("-- | -- | -- | --\n")

	for _, r := range products {
		b.WriteString(fmt.Sprintf("[%s](%s/%s.md) | %s | %s | %d\n", ncloud.StringValue(r.ProductDescription), productPath, url.QueryEscape(ncloud.StringValue(r.ProductName)), ncloud.StringValue(r.ProductCode), ncloud.StringValue(r.ProductType.CodeName), ncloud.Int64Value(r.BaseBlockStorageSize)/GIGA_BYTE))
	}

	return b.String()
}

func createMarkdownProducts(title string, products []*vserver.Product) string {
	var b bytes.Buffer
	b.WriteString("### " + title + "\n\n")
	b.WriteString("Description | Product code | Type\n")
	b.WriteString("-- | -- | --\n")

	for _, r := range products {
		b.WriteString(fmt.Sprintf("%s | %s | %s\n", ncloud.StringValue(r.ProductDescription), ncloud.StringValue(r.ProductCode), ncloud.StringValue(r.ProductType.Code)))
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
