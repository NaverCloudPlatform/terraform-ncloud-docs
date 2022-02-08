package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
)

const GIGA_BYTE = 1024 * 1024 * 1024

func main() {
	apiKeys := ncloud.Keys()
	// apiKeys := &ncloud.APIKey{
	// 	AccessKey: "YOUR_ACCESS_KEY",
	// 	SecretKey: "YOUR_SECRET_KEY",
	// }
	client := vserver.NewAPIClient(vserver.NewConfiguration(apiKeys))
	images := GetServerImageProductList(client)

	createFile(createMarkdown("Server Image Products", images), "docs/images.md")

	var wg sync.WaitGroup
	for _, r := range images {
		wg.Add(1)
		go func(r *vserver.Product) {
			defer wg.Done()
			products := GetServerProductList(client, r)
			createFile(createMarkdown(fmt.Sprintf("Server Products (%s)", *r.ProductDescription), products), "docs/products/"+*r.ProductName+".md")
		}(r)
	}

	wg.Wait()
}

func createMarkdown(title string, products []*vserver.Product) string {
	var b bytes.Buffer
	b.WriteString("### " + title + "\n")
	b.WriteString("Code | Description | O/S | B/S Size(GB)\n")
	b.WriteString("-- | -- | -- | --\n")

	for _, r := range products {
		b.WriteString(fmt.Sprintf("%s | %s | %s | %d\n", ncloud.StringValue(r.ProductCode), ncloud.StringValue(r.ProductDescription), ncloud.StringValue(r.ProductType.CodeName), ncloud.Int64Value(r.BaseBlockStorageSize)/GIGA_BYTE))
	}

	return b.String()
}

func GetServerImageProductList(client *vserver.APIClient) []*vserver.Product {
	req := vserver.GetServerImageProductListRequest{
		RegionCode: ncloud.String("KR"),
	}

	if r, err := client.V2Api.GetServerImageProductList(&req); err != nil {
		log.Println(err)
	} else {
		return r.ProductList

	}
	return nil
}

func GetServerProductList(client *vserver.APIClient, product *vserver.Product) []*vserver.Product {
	req := vserver.GetServerProductListRequest{
		RegionCode:             ncloud.String("KR"),
		ServerImageProductCode: ncloud.String(*product.ProductCode),
	}

	if r, err := client.V2Api.GetServerProductList(&req); err != nil {
		log.Println(err)
	} else {
		return r.ProductList
	}

	return nil
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
