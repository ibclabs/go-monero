package walletrpc

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/rpc/v2/json2"
)

// New returns a new monero-wallet-rpc client.
func New(cfg Config) *client {
	cl := &client{
		addr:    cfg.Address,
		headers: cfg.CustomHeaders,
		httpcl:  http.DefaultClient,
	}

	if cfg.Transport != nil {
		cl.httpcl = &http.Client{Transport: cfg.Transport}
	}

	return cl
}

type client struct {
	httpcl  *http.Client
	addr    string
	headers map[string]string
}

func (c *client) do(method string, in, out interface{}) error {
	payload, err := json2.EncodeClientRequest(method, in)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.addr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpcl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %v", resp.StatusCode)
	}

	if out == nil {
		return json2.DecodeClientResponse(resp.Body, new(json2.EmptyResponse))
	}
	return json2.DecodeClientResponse(resp.Body, out)
}

func (c *client) GetBalance() (uint64, uint64, error) {
	jd := struct {
		Balance         uint64 `json:"balance"`
		UnlockedBalance uint64 `json:"unlocked_balance"`
	}{}
	err := c.do("getbalance", nil, &jd)
	return jd.Balance, jd.UnlockedBalance, err
}

func (c *client) GetAddress() (string, error) {
	jd := struct {
		Address string `json:"address"`
	}{}
	err := c.do("getaddress", nil, &jd)
	return jd.Address, err
}

func (c *client) GetHeight() (uint64, error) {
	jd := struct {
		Height uint64 `json:"height"`
	}{}
	err := c.do("getheight", nil, &jd)
	return jd.Height, err
}

func (c *client) Transfer(req TransferRequest) (resp TransferResponse, err error) {
	err = c.do("transfer", &req, &resp)
	return
}

func (c *client) TransferSplit(req TransferRequest) (resp TransferSplitResponse, err error) {
	err = c.do("transfer_split", &req, &resp)
	return
}

func (c *client) SweepDust() ([]string, error) {
	jd := struct {
		TxHashList []string `json:"tx_hash_list"`
	}{}
	err := c.do("sweep_dust", nil, &jd)
	return jd.TxHashList, err
}

func (c *client) SweepAll(req SweepAllRequest) (resp SweepAllResponse, err error) {
	err = c.do("sweep_all", &req, &resp)
	return
}

func (c *client) Store() error {
	return c.do("store", nil, nil)
}

func (c *client) GetPayments(id string) ([]Payment, error) {
	jin := struct {
		PaymentID string `json:"payment_id"`
	}{
		id,
	}
	jd := struct {
		Payments []Payment `json:"payments"`
	}{}

	err := c.do("get_payments", &jin, &jd)
	return jd.Payments, err
}

func (c *client) GetBulkPayments(payments []string, minHeight uint) ([]Payment, error) {
	jin := struct {
		PaymentIDs     []string `json:"payment_ids"`
		MinBlockHeight uint     `json:"min_block_height"`
	}{
		payments,
		minHeight,
	}
	jd := struct {
		Payments []Payment `json:"payments"`
	}{}
	err := c.do("get_bulk_payments", &jin, &jd)
	return jd.Payments, err
}

func (c *client) GetTransfers(req GetTransfersRequest) (resp GetTransfersResponse, err error) {
	err = c.do("get_transfers", &req, &resp)
	return
}

func (c *client) GetTransferByTxID(tx string) (transfer Transfer, err error) {
	jin := struct {
		TxID string `json:"txid"`
	}{tx}

	jd := struct {
		Transfer *Transfer `json:"transfer"`
	}{}

	err = c.do("get_transfer_by_txid", &jin, &jd)
	if jd.Transfer != nil {
		transfer = *jd.Transfer
	}

	return
}

func (c *client) IncomingTransfers(transfer GetTransferType) ([]IncTransfer, error) {
	jin := struct {
		TransferType GetTransferType `json:"transfer_type"`
	}{
		transfer,
	}
	jd := struct {
		Transfers []IncTransfer `json:"transfers"`
	}{}

	err := c.do("incoming_transfers", &jin, &jd)
	return jd.Transfers, err
}

func (c *client) QueryKey(keytype QueryKeyType) (key string, err error) {
	jin := struct {
		KeyType QueryKeyType `json:"key_type"`
	}{
		keytype,
	}
	jd := struct {
		Key string `json:"key"`
	}{}
	err = c.do("query_key", &jin, &jd)
	if err != nil {
		return
	}
	key = jd.Key
	return
}

func (c *client) MakeIntegratedAddress(paymentid string) (integratedaddr string, err error) {
	jin := struct {
		PaymentID string `json:"payment_id"`
	}{
		paymentid,
	}
	jd := struct {
		Address string `json:"integrated_address"`
	}{}
	err = c.do("make_integrated_address", &jin, &jd)
	if err != nil {
		return
	}
	integratedaddr = jd.Address
	return
}

func (c *client) SplitIntegratedAddress(integratedaddr string) (paymentid, address string, err error) {
	jin := struct {
		IntegratedAddress string `json:"integrated_address"`
	}{
		integratedaddr,
	}
	jd := struct {
		Address   string `json:"standard_address"`
		PaymentID string `json:"payment_id"`
	}{}
	err = c.do("split_integrated_address", &jin, &jd)
	if err != nil {
		return
	}
	paymentid = jd.PaymentID
	address = jd.Address
	return
}

func (c *client) StopWallet() error {
	return c.do("stop_wallet", nil, nil)
}

func (c *client) MakeURI(req URIDef) (uri string, err error) {
	jd := struct {
		URI string `json:"uri"`
	}{}
	err = c.do("make_uri", &req, &jd)
	if err != nil {
		return
	}
	uri = jd.URI
	return
}

func (c *client) ParseURI(uri string) (parsed *URIDef, err error) {
	jin := struct {
		URI string `json:"uri"`
	}{
		uri,
	}
	parsed = &URIDef{}
	err = c.do("parse_uri", &jin, parsed)
	if err != nil {
		return nil, err
	}
	return
}

func (c *client) RescanBlockchain() error {
	return c.do("rescan_blockchain", nil, nil)
}

func (c *client) SetTxNotes(txids, notes []string) error {
	jin := struct {
		TxIDs []string `json:"txids"`
		Notes []string `json:"notes"`
	}{
		txids,
		notes,
	}
	return c.do("set_tx_notes", &jin, nil)
}

func (c *client) GetTxNotes(txids []string) (notes []string, err error) {
	jin := struct {
		TxIDs []string `json:"txids"`
	}{
		txids,
	}
	jd := struct {
		Notes []string `json:"notes"`
	}{}
	err = c.do("get_tx_notes", &jin, &jd)
	if err != nil {
		return nil, err
	}
	notes = jd.Notes
	return
}

func (c *client) Sign(data string) (signature string, err error) {
	jin := struct {
		Data string `json:"data"`
	}{
		data,
	}
	jd := struct {
		Signature string `json:"signature"`
	}{}
	err = c.do("sign", &jin, &jd)
	if err != nil {
		return "", err
	}
	signature = jd.Signature
	return
}

func (c *client) Verify(data, address, signature string) (good bool, err error) {
	jin := struct {
		Data      string `json:"data"`
		Address   string `json:"address"`
		Signature string `json:"signature"`
	}{
		data,
		address,
		signature,
	}
	jd := struct {
		Good bool `json:"good"`
	}{}
	err = c.do("verify", &jin, &jd)
	if err != nil {
		return false, err
	}
	good = jd.Good
	return
}

func (c *client) ExportKeyImages() ([]SignedKeyImage, error) {
	jd := struct {
		SignedKeyImages []SignedKeyImage `json:"signed_key_images"`
	}{}
	err := c.do("export_key_images", nil, &jd)
	return jd.SignedKeyImages, err
}

func (c *client) ImportKeyImages(images []SignedKeyImage) (resp ImportKeyImageResponse, err error) {
	jin := struct {
		SignedKeyImages []SignedKeyImage `json:"signed_key_images"`
	}{
		images,
	}
	resp = ImportKeyImageResponse{}
	err = c.do("import_key_images", &jin, &resp)
	return
}

func (c *client) GetAddressBook(indexes []uint64) ([]AddressBookEntry, error) {
	jin := struct {
		Indexes []uint64 `json:"entries"`
	}{
		indexes,
	}
	jd := struct {
		Entries []AddressBookEntry `json:"entries"`
	}{}
	err := c.do("get_address_book", &jin, &jd)
	return jd.Entries, err
}

func (c *client) AddAddressBook(entry AddressBookEntry) (index uint64, err error) {
	jd := struct {
		Index uint64 `json:"index"`
	}{}
	err = c.do("add_address_book", &entry, &jd)
	if err != nil {
		return 0, err
	}
	index = jd.Index
	return
}

func (c *client) DeleteAddressBook(index uint64) error {
	jin := struct {
		Index uint64 `json:"index"`
	}{
		index,
	}
	return c.do("delete_address_book", &jin, nil)
}

func (c *client) RescanSpent() error {
	return c.do("rescan_spent", nil, nil)
}

func (c *client) StartMining(threads uint, background, ignorebattery bool) error {
	jin := struct {
		Threads       uint `json:"threads_count"`
		Background    bool `json:"do_background_mining"`
		IgnoreBattery bool `json:"ignore_battery"`
	}{
		threads,
		background,
		ignorebattery,
	}
	return c.do("start_mining", &jin, nil)
}

func (c *client) StopMining() error {
	return c.do("stop_mining", nil, nil)
}

func (c *client) GetLanguages() ([]string, error) {
	jd := struct {
		Languages []string `json:"languages"`
	}{}
	err := c.do("get_languages", nil, &jd)
	if err != nil {
		return nil, err
	}
	return jd.Languages, err
}

func (c *client) CreateWallet(filename, password, language string) error {
	jin := struct {
		Filename string `json:"filename"`
		Password string `json:"password"`
		Language string `json:"language"`
	}{
		filename,
		password,
		language,
	}
	return c.do("create_wallet", &jin, nil)
}

func (c *client) OpenWallet(filename, password string) error {
	jin := struct {
		Filename string `json:"filename"`
		Password string `json:"password"`
	}{
		filename,
		password,
	}
	return c.do("open_wallet", &jin, nil)
}
