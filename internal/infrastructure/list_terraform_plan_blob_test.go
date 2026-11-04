package infrastructure

import (
	"os"
	"testing"

	"github.com/samber/lo"
)

// TestListBlob tests the ListPlanAzureStorageBlob function.
// Requires Azurite to be running locally.
// docker pull mcr.microsoft.com/azure-storage/azurite
// docker run -d -p 10000:10000 -p 10001:10001 -p 10002:10002 --restart unless-stopped mcr.microsoft.com/azure-storage/azurite azurite --skipApiVersionCheck
func TestListBlob(t *testing.T) {
	os.Setenv("AzureWebJobsStorage", "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;")
	if err := CreatePlanAzureStorageBlob("1", []byte("binary"), []byte("lock"), []byte("configuration")); err != nil {
		t.Errorf("TestSaveBlob() error = %v", err)
	}

	if ids, err := ListPlanAzureStorageBlob(0); err != nil {
		t.Errorf("TestListBlob() error = %v", err)
	} else {
		if !lo.Contains(ids, "1") {
			t.Errorf("TestListBlob() ids = %v", ids)
		}
	}
}
