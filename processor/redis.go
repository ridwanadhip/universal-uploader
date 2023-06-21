package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ridwanadhip/universal-uploader/config"
	"github.com/ridwanadhip/universal-uploader/util"
)

const (
	KeyColumn   = "key"
	ValueColumn = "value"
	TTLColumn   = "ttl"
)

type redisImplementation struct {
	input       *config.Input
	target      *config.Target
	client      *redis.Client
	verboseMode bool
}

func NewRedisImplementation(input *config.Input, target *config.Target, verboseMode bool) (*redisImplementation, error) {
	host := fmt.Sprintf("%s:%d", target.Host, target.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: target.Password,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &redisImplementation{input, target, client, verboseMode}, nil
}

func (impl *redisImplementation) Close() {
	if impl.client != nil {
		impl.client.Close()
	}
}

func (impl *redisImplementation) DryRun(data [][]string) error {
	// TODO: implement dry run
	return fmt.Errorf("not implemented")
}

func (impl *redisImplementation) Process(data [][]string) error {
	fields := impl.target.Fields

	if err := impl.validateFields(); err != nil {
		return err
	}

	newRows := []map[string]string{}
	for i := range data {
		row := map[string]string{}
		for j := range fields {
			f := &fields[j]

			fval := *f.Value
			for _, ref := range f.References {
				refID := util.RemoveToken(ref)

				// if not exists then treat missing reference value as empty string
				replacer := ""
				if j, ok := impl.input.FieldsIndexMap[refID]; ok {
					replacer = data[i][j]
				}

				fval = strings.ReplaceAll(fval, ref, replacer)
			}

			row[f.Name] = fval
		}

		newRows = append(newRows, row)
	}

	// log generated values
	if impl.verboseMode {
		fmt.Printf("[Target Redis ID: %s] %s\n", impl.target.ID, util.Jsonify(newRows))
	}

	for _, row := range newRows {
		key := row[KeyColumn]
		value := row[ValueColumn]
		rawTTL := row[TTLColumn]

		ttl, err := strconv.ParseInt(rawTTL, 10, 64)
		if err != nil {
			return fmt.Errorf("unknown TTL value: %s", err)
		}

		err = impl.client.Set(context.Background(), key, value, time.Duration(ttl)*time.Second).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (impl *redisImplementation) validateFields() error {
	_, exists := impl.target.FieldsIDMap[KeyColumn]
	if !exists {
		return fmt.Errorf("missing field in config: %s", KeyColumn)
	}

	_, exists = impl.target.FieldsIDMap[ValueColumn]
	if !exists {
		return fmt.Errorf("missing field in config: %s", ValueColumn)
	}

	_, exists = impl.target.FieldsIDMap[TTLColumn]
	if !exists {
		return fmt.Errorf("missing field in config: %s", TTLColumn)
	}

	return nil
}
