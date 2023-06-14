package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ipfs-force-community/droplet/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestRetrievalByPiece(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDri := t.TempDir()
	cfg := config.DefaultMarketConfig
	cfg.Home.HomeDir = tmpDri
	cfg.PieceStorage.Fs = []*config.FsPieceStorage{
		{
			Name:     "test",
			ReadOnly: false,
			Path:     tmpDri,
		},
	}
	assert.NoError(t, config.SaveConfig(cfg))

	pieceStr := "baga6ea4seaqpzcr744w2rvqhkedfqbuqrbo7xtkde2ol6e26khu3wni64nbpaeq"
	buf := &bytes.Buffer{}
	f, err := os.Create(filepath.Join(tmpDri, pieceStr))
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		buf.WriteString("TEST TEST\n")
	}
	_, err = f.Write(buf.Bytes())
	assert.NoError(t, err)
	assert.NoError(t, f.Close())

	s, err := newServer(tmpDri)
	assert.NoError(t, err)
	port := "34897"
	startHTTPServer(ctx, t, port, s)

	url := fmt.Sprintf("http://127.0.0.1:%s/piece/%s", port, pieceStr)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close() // nolint

	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, buf.Bytes(), data)
}

func startHTTPServer(ctx context.Context, t *testing.T, port string, s *server) {
	mux := http.DefaultServeMux
	mux.HandleFunc("/piece/", s.retrievalByPieceCID)
	ser := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	go func() {
		select {
		case <-ctx.Done():
			assert.NoError(t, ser.Shutdown(context.TODO()))
		default:
		}

		assert.NoError(t, ser.ListenAndServe())
	}()
	// wait serve up
	time.Sleep(time.Second * 2)
}
