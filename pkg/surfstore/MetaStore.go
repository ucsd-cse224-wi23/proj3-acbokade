package surfstore

import (
	context "context"
	"sync"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type MetaStore struct {
	FileMetaMap    map[string]*FileMetaData
	BlockStoreAddr string
	rwMutex sync.RWMutex
	UnimplementedMetaStoreServer
}

// Returns mapping of files and its metadata (version, filename and hashlist)
func (m *MetaStore) GetFileInfoMap(ctx context.Context, _ *emptypb.Empty) (*FileInfoMap, error) {
	// Aqcuire read lock
	m.rwMutex.RLock()
	var fileInfoMap *FileInfoMap = &FileInfoMap{FileInfoMap: m.FileMetaMap}
	m.rwMutex.RUnlock()
	return fileInfoMap, nil
}

func (m *MetaStore) UpdateFile(ctx context.Context, fileMetaData *FileMetaData) (*Version, error) {
	fileName := fileMetaData.Filename
	fileVersion := fileMetaData.Version
	// Acquire read lock
	m.rwMutex.RLock()
	_, exists := m.FileMetaMap[fileName]
	m.rwMutex.RUnlock()
	if exists {
		m.rwMutex.RLock()
		curFileVersion := m.FileMetaMap[fileName].Version
		m.rwMutex.RLock()
		// Replace the metadata only if the version is 1 greater than current file version
		if fileVersion == 1+curFileVersion {
			m.rwMutex.Lock()
			m.FileMetaMap[fileName] = fileMetaData
			m.rwMutex.Unlock()
		} else {
			// Else send version -1 to the client
			return &Version{Version: -1}, nil
		}
	} else {
		m.rwMutex.Lock()
		m.FileMetaMap[fileName] = fileMetaData
		m.rwMutex.Unlock()
	}
	return &Version{Version: fileVersion}, nil
}

func (m *MetaStore) GetBlockStoreAddr(ctx context.Context, _ *emptypb.Empty) (*BlockStoreAddr, error) {
	var blockStoreAddr *BlockStoreAddr = &BlockStoreAddr{Addr: m.BlockStoreAddr}
	return blockStoreAddr, nil
}

// This line guarantees all method for MetaStore are implemented
var _ MetaStoreInterface = new(MetaStore)

func NewMetaStore(blockStoreAddr string) *MetaStore {
	return &MetaStore{
		FileMetaMap:    map[string]*FileMetaData{},
		BlockStoreAddr: blockStoreAddr,
	}
}
