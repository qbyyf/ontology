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

package utils

import (
	"bytes"

	"github.com/qbyyf/ontology/common"
	"github.com/qbyyf/ontology/common/serialization"
	cstates "github.com/qbyyf/ontology/core/states"
	"github.com/qbyyf/ontology/errors"
	"github.com/qbyyf/ontology/smartcontract/service/native"
	"github.com/qbyyf/ontology/smartcontract/storage"
)

func GetStorageItem(cacheDB *storage.CacheDB, key []byte) (*cstates.StorageItem, error) {
	store, err := cacheDB.Get(key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageItem] storage error!")
	}
	if store == nil {
		return nil, nil
	}
	item := new(cstates.StorageItem)
	err = item.Deserialization(common.NewZeroCopySource(store))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageItem] instance doesn't StorageItem!")
	}
	return item, nil
}

func GetNativeTokenBalance(cacheDB *storage.CacheDB, key []byte) (cstates.NativeTokenBalance, error) {
	item, err := GetStorageItem(cacheDB, key)
	if err != nil {
		return cstates.NativeTokenBalance{}, err
	}
	if item == nil {
		return cstates.NativeTokenBalance{}, nil
	}

	return cstates.NativeTokenBalanceFromStorageItem(item)
}

func GetStorageUInt32(cacheDB *storage.CacheDB, key []byte) (uint32, error) {
	item, err := GetStorageItem(cacheDB, key)
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, nil
	}
	v, err := serialization.ReadUint32(bytes.NewBuffer(item.Value))
	if err != nil {
		return 0, err
	}
	return v, nil
}

func GenUInt64StorageItem(value uint64) *cstates.StorageItem {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(value)
	return &cstates.StorageItem{Value: sink.Bytes()}
}

func GenUInt32StorageItem(value uint32) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	serialization.WriteUint32(bf, value)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func PutBytes(native *native.NativeService, key []byte, value []byte) {
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
}

func GetStorageVarBytes(native *native.NativeService, key []byte) ([]byte, error) {
	item, err := GetStorageItem(native.CacheDB, key)
	if err != nil {
		return []byte{}, err
	}
	if item == nil {
		return []byte{}, nil
	}
	v, err := serialization.ReadVarBytes(bytes.NewBuffer(item.Value))
	if err != nil {
		return []byte{}, err
	}
	return v, nil
}

func GenVarBytesStorageItem(value []byte) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	serialization.WriteVarBytes(bf, value)
	return &cstates.StorageItem{Value: bf.Bytes()}
}
