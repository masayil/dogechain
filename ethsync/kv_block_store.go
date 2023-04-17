package ethsync

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dogechain-lab/dogechain/types"
	"github.com/hashicorp/go-hclog"
)

var (
	blkFormat     = "/bk/%s"
	blkHashFormat = "/bk//hash/%s"
	txFormat      = "/tx/%s"
	rtFormat      = "/rt/%s"
)

type BlockStore struct {
	kvc                 *KvSync
	pub                 *PubKv
	dlc                 *DistributedLockKv
	logger              hclog.Logger
	currentBlkHeightHex string
	isUpdate            bool
}

func NewBlockStore(logger hclog.Logger, kvc *KvSync, pub *PubKv, dlc *DistributedLockKv) *BlockStore {

	//get currentBlkHeightHex
	currentBlkHeightHexFromKV, _ := kvc.GetInTerm(context.Background(), "/ht/latest")
	return &BlockStore{
		kvc:                 kvc,
		pub:                 pub,
		dlc:                 dlc,
		logger:              logger.Named("blockstore"),
		currentBlkHeightHex: currentBlkHeightHexFromKV,
		isUpdate:            false,
	}
}

func (b *BlockStore) ToReceipt(block *types.Block, receipts []*types.Receipt) {

	if b.isUpdate {
		for index, _ := range block.Transactions {
			receiptRes := toReceiptJSONRPC(receipts, block, int64(index))

			for _, log := range receiptRes.Logs {
				logText, _ := log.Data.MarshalText()
				// b.logger.Info(fmt.Sprintf("ankr_write_receipts_log_db is %s, hash is %s", logText, receiptRes.TxHash.String()))
				if len(logText) >= 10 {
					if string(logText)[0:10] != "0x00000000" {
						b.logger.Info(fmt.Sprintf("ankr_write_receipts_log_wrong is %s, hash is %s", logText, receiptRes.TxHash.String()))

						filePath := "/tmp/dogechainfail.txt"
						file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
						if err != nil {
							fmt.Println("open dogechainfail", err)
						}
						defer file.Close()
						write := bufio.NewWriter(file)
						write.WriteString(fmt.Sprintf("ankr_write_receipts_log_wrong is %s, hash is %s \r\n", logText, receiptRes.TxHash.String()))
						write.Flush()
					}
				}
			}

			receiptJson, _ := json.Marshal(receiptRes)
			// b.logger.Info(fmt.Sprintf("ankr receipt json is %s", receiptJson))
			if err := b.kvc.SetInKV(context.Background(), fmt.Sprintf(rtFormat, receiptRes.TxHash), receiptJson, b.logger); err != nil {
				b.logger.Error(err.Error())
			}
			// pub receipt
			for _, rtLog := range receiptRes.Logs {
				logbs, _ := json.Marshal(rtLog)
				if b.pub.PublishTopicInTemp(context.Background(), b.pub.topicLogsKey, logbs) != nil {
					b.logger.Error(fmt.Sprintf("ankr pub new logs failed, tx is %s", rtLog.TxHash))
				}
			}
		}
	}

}

func (b *BlockStore) StoreBlock(block *types.Block) {

	b.logger.Info("ankr start StoreBlock", block.Number())
	b.logger.Info("ankr start StoreBlock isUpdate", b.isUpdate)

	if b.isUpdate {

		ctx := context.Background()
		blkHeightHex := fmt.Sprintf("0x%x", block.Number())
		// process block
		blockRes := toBlock(block, false)
		blockJson, _ := json.Marshal(blockRes)
		// b.logger.Info(fmt.Sprintf("ankr block key is %s， ankr block json is %s", blkFormat, blockJson))
		if err := b.kvc.SetInKV(ctx, fmt.Sprintf(blkFormat, blkHeightHex), blockJson, b.logger); err != nil {
			b.logger.Error(err.Error())
		}

		// process tx
		if len(block.Transactions) != 0 {
			for index, tx := range block.Transactions {
				txRes := toTransaction(tx, argUintPtr(block.Number()), argHashPtr(block.Hash()), &index)
				txJson, _ := json.Marshal(txRes)
				// b.logger.Info(fmt.Sprintf("ankr tx json is %s", txJson))
				if err := b.kvc.SetInKV(ctx, fmt.Sprintf(txFormat, tx.Hash()), txJson, b.logger); err != nil {
					b.logger.Error(err.Error())
				}
			}
		}

		b.kvc.SetTopicInTemp(ctx, blkHeightHex)

		b.kvc.PublishTopicInTemp(ctx, blkHeightHex)

	}

}

func (b *BlockStore) PublishTopic(ctx context.Context, block *types.Block) {
	var (
		blkHeightHex     string
		number           uint64
		currentBlkHeight uint64
	)

	number = block.Number()
	blkHeightHex = fmt.Sprintf("0x%x", number)
	// log.Info(fmt.Sprintf("ankr currentBlkHeightHex is %s", b.currentBlkHeightHex))
	currentBlkHeight, _ = strconv.ParseUint(strings.TrimPrefix(b.currentBlkHeightHex, "0x"), 16, 64)
	// log.Info(fmt.Sprintf("ankr currentBlkHeight is %d", currentBlkHeight))
	// log.Info(fmt.Sprintf("ankr number is %d", number))
	if number > currentBlkHeight {

		//get currentBlkHeightHex
		currentBlkHeightHexFromKVPub, _ := b.kvc.GetInTerm(context.Background(), "/ht/latest")
		currentBlkHeightPub, _ := strconv.ParseUint(strings.TrimPrefix(currentBlkHeightHexFromKVPub, "0x"), 16, 64)

		if number >= currentBlkHeightPub {

			b.logger.Info(fmt.Sprintf("ankr currentBlkHeightPub is %d, number is %d", currentBlkHeightPub, number))

			// set isUpdate
			b.isUpdate = true

			// set blk height

			res, err := b.dlc.Lock()
			if err != nil {
				b.logger.Error(fmt.Sprintf("ankr dlc_lock fail, err is %s", err.Error()))
			}
			if res {
				b.kvc.SetTopicInTemp(ctx, blkHeightHex)
				b.kvc.PublishTopicInTemp(ctx, blkHeightHex)
				b.logger.Info(fmt.Sprintf("ankr dlc_lock success, blk is %s", blkHeightHex))
			}
			// resInt, err := b.dlc.UnLock()
			if err != nil {
				b.logger.Error(fmt.Sprintf("ankr dlc_unlock fail, err is %s", err.Error()))
			}
			// b.logger.Info(fmt.Sprintf("ankr dlc_unlock success, blk is %s, redInt is %d", blkHeightHex, resInt))

			// set header height
			header, err := json.Marshal(block.Header)
			if err != nil {
				b.logger.Error("ankr json Marshal block fail, block is ", blkHeightHex)
			}
			b.pub.PublishTopicInTemp(ctx, b.pub.topicHeaderKey, header)
			return
		} else {
			b.isUpdate = false
		}
	}
}
