package config

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/compression"
	"github.com/libp2p/go-libp2p-core/host"

	csms "github.com/libp2p/go-conn-compression-multistream"
)

// CompC is a security transport constructor.
type CompC func(h host.Host) (compression.CompressedTransport, error)

// MsCompC is a tuple containing a security transport constructor and a protocol
// ID.
type MsCompC struct {
	CompC
	ID string
}

var compressionArgTypes = newArgTypeSet(hostType, networkType, peerIDType, pstoreType)

// CompressionConstructor creates a security constructor from the passed parameter
// using reflection.
func CompressionConstructor(compressor interface{}) (CompC, error) {
	// Already constructed?
	if t, ok := compressor.(compression.CompressedTransport); ok {
		return func(_ host.Host) (compression.CompressedTransport, error) {
			return t, nil
		}, nil
	}

	ctor, err := makeConstructor(compressor, compressionType, compressionArgTypes)
	if err != nil {
		return nil, err
	}
	return func(h host.Host) (compression.CompressedTransport, error) {
		t, err := ctor(h, nil, nil)
		if err != nil {
			return nil, err
		}
		return t.(compression.CompressedTransport), nil
	}, nil
}

func makeCompressionTransport(h host.Host, tpts []MsCompC) (compression.CompressedTransport, error) {
	compMux := new(csms.Transport)
	transportSet := make(map[string]struct{}, len(tpts))
	for _, tptC := range tpts {
		if _, ok := transportSet[tptC.ID]; ok {
			return nil, fmt.Errorf("duplicate compression transport: %s", tptC.ID)
		}
		transportSet[tptC.ID] = struct{}{}
	}
	for _, tptC := range tpts {
		tpt, err := tptC.CompC(h)
		if err != nil {
			return nil, err
		}
		compMux.AddTransport(tptC.ID, tpt)
	}
	return compMux, nil
}
