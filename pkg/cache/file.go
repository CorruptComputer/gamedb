package cache

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/djherbis/fscache"
	"go.uber.org/zap"
)

func GetSetCache(name string, ttl time.Duration, retrieve func() (interface{}, error), val interface{}) (err error) {

	if ttl == 0 {
		ttl = time.Hour * 24 * 365
	}

	c, err := fscache.New("./cache", 0755, ttl)
	if err != nil {
		return err
	}

	reader, writer, err := c.Get(name)
	if err != nil {
		return err
	}

	defer func() {
		err = reader.Close()
		if err != nil {
			zap.S().Error(err)
		}
	}()

	// Read from cache
	if writer == nil {
		dec := gob.NewDecoder(reader)
		return dec.Decode(val)
	}

	zap.L().Info("Saving " + name + " to cache")

	// Write to cache
	defer func() {
		err = writer.Close()
		if err != nil {
			zap.S().Error(err)
		}
	}()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	i, err := retrieve()
	if err != nil {
		return err
	}

	err = encoder.Encode(i)
	if err != nil {
		return err
	}

	// Save to cache
	_, err = writer.Write(buf.Bytes())
	return err
}
