/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ledger

import (
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/qbyyf/ontology/core/store"
	"github.com/qbyyf/ontology/core/store/ledgerstore"
	"github.com/qbyyf/ontology/core/types"
)

var DefLedger *Ledger

type Ledger struct {
	store.LedgerStore
}

func (self *Ledger) GetStore() store.LedgerStore {
	return self.LedgerStore
}

func InitLedger(dataDir string, stateHashHeight uint32, defaultBookkeeper []keypair.PublicKey,
	genesisBlock *types.Block) (*Ledger, error) {
	ldgStore, err := ledgerstore.NewLedgerStore(dataDir, stateHashHeight)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}

	err = ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock, defaultBookkeeper)
	if err != nil {
		return nil, err
	}

	return &Ledger{
		LedgerStore: ldgStore,
	}, nil
}
