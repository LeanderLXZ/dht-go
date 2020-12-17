package src

import(
	"hash"
)

type file interface {
	Get(string) ([]byte， error)
	Set(string, string) error
	Delete(string) error
	Between([]byte, []byte) ([]*KeyValuePair, error)
	MDelete(...string) error 
}

type DataHash struct {
	data map[string]string
	Hash func() hash.Hash 
}

func NewDataHash(hashFunc func() hash.Hash) file {
	return &DataHash{
		data: make(map[string]string),
		Hash: hashFunc,
	}
}

func (dh *DataHash) hashKey(key string) ([]byte, error) {
	h := dh.Hash()
	if _, err := h.Write([]byte(key)); err != nil {
		return nil, err
	}
	val := h.Sum(nil)
	return val, nil
}

func (dh *DataHash) AddKey(key, value string) error {
	dh.data[key] = value
	return nil
}

func (dh *DataHash) GetValue(key string) ([]byte, error) {
	value, res := dh.data[key]
	if !res {
		return nil, ERR_KEY_NOT_FOUND
	}
	return []byte(value), nil
}

func (dh *DataHash) DeleteKey(key string) error {
	delete(dh.data, key)
	return nil
}

func (dh *DataHash) DeleteKeys(keys ...string) error {
	for _, k := range keys {
		delete(dh.data, k)
	}
	return nil
}

func (dh *DataHash) Between(from []byte, to []byte) ([]*KeyValuePair, error) {
	values := make([]*KeyValuePair, 0, 10)
	for k, v := range dh.data {
		hashedKey, err := dh.hashKey(k)
		if err != nil {
			continue
		}
		if betweenRightIncl(hashedKey, from, to) {
			pair := &KeyValuePair{
				Key:   k,
				Value: v,
			}
			values = append(values, pair)
		}
	}
	return values, nil
}


