package handler

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/0chain/blobber/code/go/0chain.net/core/common"

	"google.golang.org/grpc/metadata"

	"github.com/0chain/blobber/code/go/0chain.net/core/encryption"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/stats"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/reference"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/allocation"

	"github.com/stretchr/testify/mock"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc"
)

func TestBlobberGRPCService_GetAllocation_Success(t *testing.T) {
	req := &blobbergrpc.GetAllocationRequest{
		Id: "something",
	}

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Id, false).Return(&allocation.Allocation{
		Tx: req.Id,
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	allocation, err := svc.GetAllocation(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, allocation.Allocation.Tx, req.Id)
}

func TestBlobberGRPCService_GetAllocation_invalidAllocation(t *testing.T) {
	req := &blobbergrpc.GetAllocationRequest{
		Id: "invalid_allocation",
	}

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Id, false).Return(nil, errors.New("some error"))

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.GetAllocation(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}

	assert.Equal(t, err.Error(), "some error")
}

func TestBlobberGRPCService_GetFileMetaData_Success(t *testing.T) {
	req := &blobbergrpc.GetFileMetaDataRequest{
		Path:       "path",
		PathHash:   "path_hash",
		AuthToken:  "testval",
		Allocation: "something",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "client",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, true).Return(&allocation.Allocation{
		ID: "allocationId",
		Tx: req.Allocation,
	}, nil)
	mockReferencePackage.On("GetReferenceFromLookupHash", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name: "test",
		Type: reference.FILE,
	}, nil)
	mockReferencePackage.On("GetCommitMetaTxns", mock.Anything, mock.Anything).Return(nil, nil)
	mockReferencePackage.On("GetCollaborators", mock.Anything, mock.Anything).Return([]reference.Collaborator{
		reference.Collaborator{
			RefID:    1,
			ClientID: "test",
		},
	}, nil)
	mockReferencePackage.On("IsACollaborator", mock.Anything, mock.Anything, mock.Anything).Return(true)
	mockStorageHandler.On("verifyAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	resp, err := svc.GetFileMetaData(ctx, req)
	if err != nil {
		t.Fatal("unexpected error")
	}

	assert.Equal(t, resp.MetaData.FileMetaData.Name, "test")
}

func TestBlobberGRPCService_GetFileMetaData_FileNotExist(t *testing.T) {
	req := &blobbergrpc.GetFileMetaDataRequest{
		Path:       "path",
		PathHash:   "path_hash",
		AuthToken:  "testval",
		Allocation: "something",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "client",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, true).Return(&allocation.Allocation{
		ID: "allocationId",
		Tx: req.Allocation,
	}, nil)
	mockReferencePackage.On("GetReferenceFromLookupHash", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("file doesnt exist"))
	mockReferencePackage.On("GetCommitMetaTxns", mock.Anything, mock.Anything).Return(nil, nil)
	mockReferencePackage.On("GetCollaborators", mock.Anything, mock.Anything).Return([]reference.Collaborator{
		reference.Collaborator{
			RefID:    1,
			ClientID: "test",
		},
	}, nil)
	mockReferencePackage.On("IsACollaborator", mock.Anything, mock.Anything, mock.Anything).Return(true)
	mockStorageHandler.On("verifyAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.GetFileMetaData(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func randString(n int) string {

	const hexLetters = "abcdef0123456789"

	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteByte(hexLetters[rand.Intn(len(hexLetters))])
	}
	return sb.String()
}

func TestBlobberGRPCService_GetFileStats_Success(t *testing.T) {
	allocationTx := randString(32)

	pubKey, _, signScheme := GeneratePubPrivateKey(t)
	clientSignature, _ := signScheme.Sign(encryption.Hash(allocationTx))

	req := &blobbergrpc.GetFileStatsRequest{
		Path:       "path",
		PathHash:   "path_hash",
		Allocation: allocationTx,
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "owner",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: clientSignature,
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, true).Return(&allocation.Allocation{
		ID:             "allocationId",
		Tx:             req.Allocation,
		OwnerID:        "owner",
		OwnerPublicKey: pubKey,
	}, nil)
	mockReferencePackage.On("GetReferenceFromLookupHash", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		ID:   123,
		Name: "test",
		Type: reference.FILE,
	}, nil)
	mockReferencePackage.On("GetFileStats", mock.Anything, int64(123)).Return(&stats.FileStats{
		NumBlockDownloads: 10,
	}, nil)
	mockReferencePackage.On("GetWriteMarkerEntity", mock.Anything, mock.Anything).Return(nil, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	resp, err := svc.GetFileStats(ctx, req)
	if err != nil {
		t.Fatal("unexpected error")
	}

	assert.Equal(t, resp.MetaData.FileMetaData.Name, "test")
	assert.Equal(t, resp.Stats.NumBlockDownloads, int64(10))
}

func TestBlobberGRPCService_GetFileStats_FileNotExist(t *testing.T) {
	req := &blobbergrpc.GetFileStatsRequest{
		Path:       "path",
		PathHash:   "path_hash",
		Allocation: "",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "owner",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, true).Return(&allocation.Allocation{
		ID:      "allocationId",
		Tx:      req.Allocation,
		OwnerID: "owner",
	}, nil)
	mockReferencePackage.On("GetReferenceFromLookupHash", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("file does not exist"))
	mockReferencePackage.On("GetFileStats", mock.Anything, int64(123)).Return(&stats.FileStats{
		NumBlockDownloads: 10,
	}, nil)
	mockReferencePackage.On("GetWriteMarkerEntity", mock.Anything, mock.Anything).Return(nil, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.GetFileStats(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlobberGRPCService_ListEntities_Success(t *testing.T) {
	req := &blobbergrpc.ListEntitiesRequest{
		Path:       "path",
		PathHash:   "path_hash",
		AuthToken:  "something",
		Allocation: "",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "client",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, true).Return(&allocation.Allocation{
		ID:             "allocationId",
		Tx:             req.Allocation,
		OwnerID:        "owner",
		AllocationRoot: "/allocationroot",
	}, nil)
	mockReferencePackage.On("GetReferenceFromLookupHash", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name: "test",
		Type: reference.FILE,
	}, nil)
	mockStorageHandler.On("verifyAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	mockReferencePackage.On("GetRefWithChildren", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name: "test",
		Type: reference.DIRECTORY,
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	resp, err := svc.ListEntities(ctx, req)
	if err != nil {
		t.Fatal("unexpected error")
	}

	assert.Equal(t, resp.AllocationRoot, "/allocationroot")
}

func TestBlobberGRPCService_ListEntities_InvalidAuthTicket(t *testing.T) {
	req := &blobbergrpc.ListEntitiesRequest{
		Path:       "path",
		PathHash:   "path_hash",
		AuthToken:  "something",
		Allocation: "",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "client",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, true).Return(&allocation.Allocation{
		ID:      "allocationId",
		Tx:      req.Allocation,
		OwnerID: "owner",
	}, nil)
	mockReferencePackage.On("GetReferenceFromLookupHash", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name: "test",
		Type: reference.FILE,
	}, nil)
	mockStorageHandler.On("verifyAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	mockReferencePackage.On("GetRefWithChildren", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name: "test",
		Type: reference.DIRECTORY,
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.ListEntities(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlobberGRPCService_GetObjectPath_Success(t *testing.T) {
	allocationTx := randString(32)

	pubKey, _, signScheme := GeneratePubPrivateKey(t)
	clientSignature, _ := signScheme.Sign(encryption.Hash(allocationTx))

	req := &blobbergrpc.GetObjectPathRequest{
		Allocation: allocationTx,
		Path:       "path",
		BlockNum:   "120",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "owner",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: clientSignature,
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, false).Return(&allocation.Allocation{
		ID:             "allocationId",
		Tx:             req.Allocation,
		OwnerID:        "owner",
		OwnerPublicKey: pubKey,
	}, nil)
	mockReferencePackage.On("GetObjectPath", mock.Anything, mock.Anything, mock.Anything).Return(&reference.ObjectPath{
		RootHash:     "hash",
		FileBlockNum: 120,
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	resp, err := svc.GetObjectPath(ctx, req)
	if err != nil {
		t.Fatal("unexpected error")
	}

	assert.Equal(t, resp.ObjectPath.RootHash, "hash")
	assert.Equal(t, resp.ObjectPath.FileBlockNum, int64(120))

}

func TestBlobberGRPCService_GetObjectPath_InvalidAllocation(t *testing.T) {
	req := &blobbergrpc.GetObjectPathRequest{
		Allocation: "",
		Path:       "path",
		BlockNum:   "120",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "owner",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, false).Return(nil, errors.New("invalid allocation"))
	mockReferencePackage.On("GetObjectPathGRPC", mock.Anything, mock.Anything, mock.Anything).Return(&blobbergrpc.ObjectPath{
		RootHash:     "hash",
		FileBlockNum: 120,
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.GetObjectPath(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlobberGRPCService_GetReferencePath_Success(t *testing.T) {
	allocationTx := randString(32)

	pubKey, _, signScheme := GeneratePubPrivateKey(t)
	clientSignature, _ := signScheme.Sign(encryption.Hash(allocationTx))

	req := &blobbergrpc.GetReferencePathRequest{
		Paths:      `["something"]`,
		Path:       "",
		Allocation: allocationTx,
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "client",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: clientSignature,
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, false).Return(&allocation.Allocation{
		ID:             "allocationId",
		Tx:             req.Allocation,
		OwnerID:        "owner",
		OwnerPublicKey: pubKey,
	}, nil)
	mockReferencePackage.On("GetReferencePathFromPaths", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name:     "test",
		Type:     reference.DIRECTORY,
		Children: []*reference.Ref{{Name: "test1", Type: reference.FILE}},
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	resp, err := svc.GetReferencePath(ctx, req)
	if err != nil {
		t.Fatal("unexpected error")
	}

	assert.Equal(t, resp.ReferencePath.MetaData.DirMetaData.Name, "test")

}

func TestBlobberGRPCService_GetReferencePath_InvalidPaths(t *testing.T) {
	allocationTx := randString(32)

	pubKey, _, signScheme := GeneratePubPrivateKey(t)
	clientSignature, _ := signScheme.Sign(encryption.Hash(allocationTx))

	req := &blobbergrpc.GetReferencePathRequest{
		Paths:      `["something"]`,
		Path:       "",
		Allocation: allocationTx,
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "client",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: clientSignature,
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, false).Return(&allocation.Allocation{
		ID:             "allocationId",
		Tx:             req.Allocation,
		OwnerID:        "owner",
		OwnerPublicKey: pubKey,
	}, nil)
	mockReferencePackage.On("GetReferencePathFromPaths", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("invalid paths"))

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.GetReferencePath(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}

	assert.Equal(t, err.Error(), "invalid paths")

}

func TestBlobberGRPCService_GetObjectTree_Success(t *testing.T) {
	allocationTx := randString(32)

	pubKey, _, signScheme := GeneratePubPrivateKey(t)
	clientSignature, _ := signScheme.Sign(encryption.Hash(allocationTx))

	req := &blobbergrpc.GetObjectTreeRequest{
		Path:       "something",
		Allocation: allocationTx,
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "owner",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: clientSignature,
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, false).Return(&allocation.Allocation{
		ID:             "allocationId",
		Tx:             req.Allocation,
		OwnerID:        "owner",
		OwnerPublicKey: pubKey,
	}, nil)
	mockReferencePackage.On("GetObjectTree", mock.Anything, mock.Anything, mock.Anything).Return(&reference.Ref{
		Name:     "test",
		Type:     reference.DIRECTORY,
		Children: []*reference.Ref{{Name: "test1", Type: reference.FILE}},
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	resp, err := svc.GetObjectTree(ctx, req)
	if err != nil {
		t.Fatal("unexpected error - " + err.Error())
	}

	assert.Equal(t, resp.ReferencePath.MetaData.DirMetaData.Name, "test")

}

func TestBlobberGRPCService_GetObjectTree_NotOwner(t *testing.T) {
	req := &blobbergrpc.GetObjectTreeRequest{
		Path:       "something",
		Allocation: "",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		common.ClientHeader:          "hacker",
		common.ClientKeyHeader:       "",
		common.ClientSignatureHeader: "",
	}))

	mockStorageHandler := &storageHandlerI{}
	mockReferencePackage := &mocks.PackageHandler{}
	mockStorageHandler.On("verifyAllocation", mock.Anything, req.Allocation, false).Return(&allocation.Allocation{
		ID:      "allocationId",
		Tx:      req.Allocation,
		OwnerID: "owner",
	}, nil)

	svc := newGRPCBlobberService(mockStorageHandler, mockReferencePackage)
	_, err := svc.GetObjectTree(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}

}