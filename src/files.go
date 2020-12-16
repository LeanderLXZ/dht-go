package chord

import(
	"hash"
)

type file interface {
	Get(string) ([]byte， error)
	Set(string, string) error
	Delete(string) error
	Between([]byte, []byte) ([]*models.KV, error)
	MDelete(...string) error 
}

type DataHash struct {
	data map[string]string
	Hash func() hash.Hash 
}

func NewDataHash(hashFunc func() hash.Hash) Storage {
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

func (dh *DataHash) Set(key, value string) error {
	dh.data[key] = value
	return nil
}

func (dh *DataHash) Get(key string) ([]byte, error) {
	value, res := dh.data[key]
	if !res {
		return nil, ERR_KEY_NOT_FOUND
	}
	return []byte(value), nil
}

func (dh *DataHash) Delete(key string) error {
	delete(dh.data, key)
	return nil
}

func (dh *DataHash) MDelete(keys ...string) error {
	for _, k := range keys {
		delete(dh.data, k)
	}
	return nil
}

func (dh *DataHash) Between(from []byte, to []byte) ([]*models.KV, error) {
	values := make([]*models.KV, 0, 10)
	for k, v := range dh.data {
		hashedKey, err := dh.hashKey(k)
		if err != nil {
			continue
		}
		if betweenRightIncl(hashedKey, from, to) {
			pair := &models.KV{
				Key:   k,
				Value: v,
			}
			values = append(values, pair)
		}
	}
	return values, nil
}


