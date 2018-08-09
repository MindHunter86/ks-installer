package boltdb

import (
	"errors"
	"reflect"

	"github.com/coreos/bbolt"
	"github.com/rs/zerolog"
)

type (
	BaseBucket struct {
		bDB     *bolt.DB
		bBucket *bolt.Bucket

		log *zerolog.Logger

		Name []byte
	}
	Bucket interface {
		Init() error
		Get(key []byte) []byte
		Put(key, value []byte) error
		Del(key []byte) error
	}
)

const (
	boltBucketKeyPrimary = uint8(iota)
	boltBucketKeyForeign
)

func NewBaseBucket(bDB *bolt.DB, log *zerolog.Logger, name []byte) *BaseBucket {
	return &BaseBucket{
		bDB:  bDB,
		log:  log,
		Name: name,
	}
}

func (m *BaseBucket) LoadOrCreate() error {
	tx, e := m.bDB.Begin(true)
	if e != nil {
		return e
	}
	defer tx.Rollback()

	if m.bBucket, e = tx.CreateBucketIfNotExists(m.Name); e != nil {
		return e
	}

	if e = tx.Commit(); e != nil {
		return e
	}

	return nil
}

// todo: adaptate for new marshal func()
func (m *BaseBucket) Put(input interface{}) (e error) {
	kvPairs, e := m.marshal(input)
	if e != nil {
		return e
	}

	var tx *bolt.Tx
	if tx, e = m.bDB.Begin(true); e != nil {
		return e
	}
	defer tx.Rollback()

	var txBucket = tx.Bucket(m.Name)
	if txBucket == nil {
		return errBoltPossibleBrokenScheme
	}

	for k, v := range kvPairs {
		if e = txBucket.Put([]byte(k), v); e != nil {
			return e
		}
	}

	return tx.Commit()
}

func (m *BaseBucket) Del(key []byte) error {
	return m.bBucket.Delete(key)
}

func (m *BaseBucket) Get(tpl interface{}, key string, where ...string) (result []interface{}, e error) {
	boltKeys, kvPairs, e := m.marshal(tpl)
	if e != nil {
		return
	}

	e = m.bDB.View(func(tx *bolt.Tx) (err error) {
		var txBucket *bolt.Bucket
		if txBucket := tx.Bucket(m.Name); txBucket == nil {
			return errBoltPossibleBrokenScheme
		}

		txCursor := txBucket.Cursor()

		for k, v := txCursor.Seek(key); k != nil && bytes.Contains(k, []byte(key); k, v = txCursor.Next() {
			append(result, &)
		}
		
	})

	return
}

/* 	e = m.bDB.View(func(tx *bolt.Tx) (err error) {
	var txBucket *bolt.Bucket
	if txBucket := tx.Bucket(m.Name); txBucket == nil {
		return errBoltPossibleBrokenScheme
	}

	txCursor := txBucket.Cursor()

	var seekedRecords
	for _, _ := txCursor.Seek(seekKey.Bytes()); k != nil && bytes.Contains(k, seekKey.Bytes()); txCursor.Next() {
		seekedRecords[k] = v
	}

	return
}) */

func (m *BaseBucket) unmarshal(input []byte) (interface{}, error) {

	return nil, nil
}

func (m *BaseBucket) marshal(input interface{}) (boltKeys map[uint8]string, kvPairs map[string]string, e error) {
	var inputType = reflect.TypeOf(input)
	var inputValue = reflect.ValueOf(input)

	if inputType.Kind() != reflect.Struct {
		return boltKeys, kvPairs, errBoltInvalidInputData
	}

	var inputFieldNum = inputType.NumField()
	for i := 0; i < inputFieldNum; i++ {
		if tagValue, ok := inputType.Field(i).Tag.Lookup("bolt"); ok {
			switch tagValue {
			case "primary_key":
				boltKeys[boltBucketKeyPrimary] = inputValue.Field(i).String()
			case "foreign_key":
				boltKeys[boltBucketKeyForeign] = inputValue.Field(i).String()
			default:
				m.log.Warn().Str("parsedTag", tagValue).Msg("An abnormal struct tag was caught")
			}
		} else {
			kvPairs[inputType.Field(i).Name] = inputValue.Field(i).String()
		}
	}

	if len(boltKeys) == 0 && len(kvPairs) == 0 {
		return boltKeys, kvPairs, errBoltAbnormalStructParse
	}

	return
}

/* func (m *BaseBucket) marshal(input interface{}) (kvPairs map[string][]byte, _ error) {
	var inputType = reflect.TypeOf(input)
	var inputValue = reflect.ValueOf(input)

	if inputType.Kind() != reflect.Struct {
		return kvPairs, errBoltInvalidInputData
	}

	var propList map[string]string
	var propPrefix string

	var inputFieldNum = inputType.NumField()
	for i := 0; i < inputFieldNum; i++ {
		if tagValue, ok := inputType.Field(i).Tag.Lookup("bolt"); ok {
			switch tagValue {
			case "primary_key":
				propPrefix = inputValue.Field(i).String() + ":" + propPrefix
			case "foreign_key":
				propPrefix = propPrefix + inputValue.Field(i).String() + ":"
			default:
				m.log.Warn().Str("parsedTag", tagValue).Msg("An abnormal struct tag was caught")
			}
		} else {
			propList[inputType.Field(i).Name] = inputValue.Field(i).String()
		}
	}

	if len(propList) == 0 && propPrefix == "" {
		return kvPairs, errBoltAbnormalStructParse
	}

	for k, v := range propList {
		kvPairs[propPrefix+k] = []byte(v)
	}

	return kvPairs, nil
}
*/
var (
	errBoltInvalidInputData     = errors.New("")
	errBoltAbnormalStructParse  = errors.New("")
	errBoltPossibleBrokenScheme = errors.New("")
)
