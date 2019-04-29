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
