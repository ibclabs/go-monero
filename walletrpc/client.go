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
	}
	if cfg.Transport == nil {
		cl.httpcl = http.DefaultClient
	} else {
		cl.httpcl = &http.Client{
			Transport: cfg.Transport,
		}
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
	if c.headers != nil {
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.httpcl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %v", resp.StatusCode)
	}

	// in theory this is only done to catch
	// any monero related errors if
	// we are not expecting any data back
	if out == nil {
		v := &json2.EmptyResponse{}
		return json2.DecodeClientResponse(resp.Body, v)
	}
	return json2.DecodeClientResponse(resp.Body, out)
}

func (c *client) GetBalance() (balance, unlockedBalance uint64, err error) {
	jd := struct {
		Balance         uint64 `json:"balance"`
		UnlockedBalance uint64 `json:"unlocked_balance"`
	}{}
	err = c.do("getbalance", nil, &jd)
	return jd.Balance, jd.UnlockedBalance, err
}

func (c *client) GetAddress() (address string, err error) {
	jd := struct {
		Address string `json:"address"`
	}{}
	err = c.do("getaddress", nil, &jd)
	if err != nil {
		return "", err
	}
	return jd.Address, nil
}

func (c *client) GetHeight() (height uint64, err error) {
	jd := struct {
		Height uint64 `json:"height"`
	}{}
	err = c.do("getheight", nil, &jd)
	if err != nil {
		return 0, err
	}
	return jd.Height, nil
}

func (c *client) Transfer(req TransferRequest) (resp *TransferResponse, err error) {
	resp = &TransferResponse{}
	err = c.do("transfer", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) TransferSplit(req TransferRequest) (resp *TransferSplitResponse, err error) {
	resp = &TransferSplitResponse{}
	err = c.do("transfer_split", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) SweepDust() (txHashList []string, err error) {
	jd := struct {
		TxHashList []string `json:"tx_hash_list"`
	}{}
	err = c.do("sweep_dust", nil, &jd)
	if err != nil {
		return nil, err
	}
	return jd.TxHashList, nil
}

func (c *client) SweepAll(req SweepAllRequest) (resp *SweepAllResponse, err error) {
	resp = &SweepAllResponse{}
	err = c.do("sweep_all", &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) Store() error {
	return c.do("store", nil, nil)
}

func (c *client) GetPayments(paymentid string) (payments []Payment, err error) {
	jin := struct {
		PaymentID string `json:"payment_id"`
	}{
		paymentid,
	}
	jd := struct {
		Payments []Payment `json:"payments"`
	}{}
	err = c.do("get_payments", &jin, &jd)
	if err != nil {
		return nil, err
	}
	return jd.Payments, nil
}

func (c *client) GetBulkPayments(paymentids []string, minblockheight uint) (payments []Payment, err error) {
	jin := struct {
		PaymentIDs     []string `json:"payment_ids"`
		MinBlockHeight uint     `json:"min_block_height"`
	}{
		paymentids,
		minblockheight,
	}
	jd := struct {
		Payments []Payment `json:"payments"`
	}{}
	err = c.do("get_bulk_payments", &jin, &jd)
	if err != nil {
		return nil, err
	}
	return jd.Payments, nil
}

func (c *client) GetTransfers(req GetTransfersRequest) (resp *GetTransfersResponse, err error) {
	resp = &GetTransfersResponse{}
	err = c.do("get_transfers", &req, resp)
	return
}

func (c *client) GetTransferByTxID(txid string) (transfer *Transfer, err error) {
	jin := struct {
		TxID string `json:"txid"`
	}{
		txid,
	}
	jd := struct {
		Transfer *Transfer `json:"transfer"`
	}{}
	err = c.do("get_transfer_by_txid", &jin, &jd)
	if err != nil {
		return
	}
	transfer = jd.Transfer
	return
}

func (c *client) IncomingTransfers(transfertype GetTransferType) (transfers []IncTransfer, err error) {
	jin := struct {
		TransferType GetTransferType `json:"transfer_type"`
	}{
		transfertype,
	}
	jd := struct {
		Transfers []IncTransfer `json:"transfers"`
	}{}
	err = c.do("incoming_transfers", &jin, &jd)
	if err != nil {
		return
	}
	transfers = jd.Transfers
	return
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

func (c *client) ExportKeyImages() (signedkeyimages []SignedKeyImage, err error) {
	jd := struct {
		SignedKeyImages []SignedKeyImage `json:"signed_key_images"`
	}{}
	err = c.do("export_key_images", nil, &jd)
	signedkeyimages = jd.SignedKeyImages
	return
}

func (c *client) ImportKeyImages(signedkeyimages []SignedKeyImage) (resp *ImportKeyImageResponse, err error) {
	jin := struct {
		SignedKeyImages []SignedKeyImage `json:"signed_key_images"`
	}{
		signedkeyimages,
	}
	resp = &ImportKeyImageResponse{}
	err = c.do("import_key_images", &jin, resp)
	if err != nil {
		return nil, err
	}
	return
}

func (c *client) GetAddressBook(indexes []uint64) (entries []AddressBookEntry, err error) {
	jin := struct {
		Indexes []uint64 `json:"entries"`
	}{
		indexes,
	}
	jd := struct {
		Entries []AddressBookEntry `json:"entries"`
	}{}
	err = c.do("get_address_book", &jin, &jd)
	if err != nil {
		return nil, err
	}
	entries = jd.Entries
	return
}

func (c *client) AddAddressBook(entry AddressBookEntry) (index uint64, err error) {
	entry.Index = 0
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

func (c *client) GetLanguages() (languages []string, err error) {
	jd := struct {
		Languages []string `json:"languages"`
	}{}
	err = c.do("get_languages", nil, &jd)
	if err != nil {
		return nil, err
	}
	languages = jd.Languages
	return
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
