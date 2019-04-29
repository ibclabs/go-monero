package walletrpc

type BlockHeaderResponse struct {
	BlockHeader BlockHeader `json:"block_header"`
	Status      string      `json:"status"`
}

type BlockHeader struct {
	Depth        uint64 `json:"depth"`
	Difficulty   uint   `json:"difficulty"`
	Hash         string `json:"hash"`
	Height       uint   `json:"height"`
	MajorVersion uint   `json:"major_version"`
	MinorVersion uint   `json:"minor_version"`
	Nonce        uint   `json:"nonce"`
	OrphanStatus bool   `json:"orphan_status"`
	PrevHash     string `json:"prev_hash"`
	Reward       uint   `json:"reward "`
	Timestamp    uint   `json:"timestamp"`
}

func (c *client) GetLastBlockHeader() (res BlockHeaderResponse, err error) {
	err = c.do("getlastblockheader", nil, &res)
	return
}

type Block struct {
	Blob        string      `json:"blob"`
	BlockHeader BlockHeader `json:"block_header"`
	Txs         []string    `json:"tx_hashes"`
	Status      string      `json:"status"`
}

func (c *client) GetBlockByHeight(height uint) (res Block, err error) {
	req := struct {
		Height uint `json:"height"`
	}{height}
	err = c.do("getblock", req, &res)
	return
}

func (c *client) GetBlockByHash(hash string) (res Block, err error) {
	req := struct {
		Hash string `json:"hash"`
	}{hash}
	err = c.do("getblock", req, &res)
	return
}
