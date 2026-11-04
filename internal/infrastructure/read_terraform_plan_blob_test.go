package infrastructure

import (
	"os"
	"testing"
)

// TestReadBlob tests the ReadPlanAzureStorageBlob function.
// Requires Azurite to be running locally.
// docker pull mcr.microsoft.com/azure-storage/azurite
// docker run -p 10000:10000 -p 10001:10001 -p 10002:10002 -d mcr.microsoft.com/azure-storage/azurite azurite --skipApiVersionCheck
func TestReadBlob(t *testing.T) {
	os.Setenv("AzureWebJobsStorage", "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;")
	if err := CreatePlanAzureStorageBlob("1", []byte("binary"), []byte("lock"), []byte("configuration")); err != nil {
		t.Errorf("TestSaveBlob() error = %v", err)
	}

	planBinary, planLock, planConfig, err := ReadPlanAzureStorageBlob("1")

	if err != nil {
		t.Errorf("TestReadBlob() error = %v", err)
	}

	if string(planBinary) != "binary" {
		t.Errorf("TestReadBlob() planBinary = %v", string(planBinary))
	}

	if string(planLock) != "lock" {
		t.Errorf("TestReadBlob() planLock = %v", string(planLock))
	}

	if string(planConfig) != "configuration" {
		t.Errorf("TestReadBlob() planConfig = %v", string(planConfig))
	}
}
