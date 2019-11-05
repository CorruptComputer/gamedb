package helpers

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/djherbis/fscache"
	"github.com/gamedb/gamedb/pkg/log"
)

func GetSetCache(name string, ttl time.Duration, retrieve func() interface{}, val interface{}) (err error) {

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
		log.Err(err)
	}()

	if writer == nil {

		// Read from cache
		dec := gob.NewDecoder(reader)
		return dec.Decode(val)

	} else {

		log.Info("Saving " + name + " to cache")

		// Write to cache
		defer func() {
			err = writer.Close()
			log.Err(err)
		}()

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		err := encoder.Encode(retrieve())
		if err != nil {
			return err
		}

		// Save to cache
		_, err = writer.Write(buf.Bytes())
		return err
	}
}
