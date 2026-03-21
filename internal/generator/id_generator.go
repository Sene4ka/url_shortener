package generator

import (
	"strings"
	"time"

	"github.com/sony/sonyflake/v2"
)

type IdGenerator struct {
	flake *sonyflake.Sonyflake
}

func NewIdGenerator(startTime time.Time, sequenceBits uint64) (*IdGenerator, error) {
	/*
		fixed 4 bits of machineId (last of int64) to ignore as fewer bits as possible
		we need 59 bits for base62 to return <= 10 symbols, but default is 63 with no sign
		can do this because we do not need a distributed uniqueness guarantee for now
		if needed later we need to change this, fine variant is to ignore 4 sequence bits
	*/
	st := sonyflake.Settings{
		BitsSequence:  int(sequenceBits),
		BitsMachineID: 4,
		TimeUnit:      time.Millisecond,
		StartTime:     startTime,
		MachineID: func() (int, error) {
			return 1, nil
		},
	}

	flake, err := sonyflake.New(st)

	if err != nil {
		return nil, err
	}

	return &IdGenerator{flake: flake}, nil
}

func (g *IdGenerator) GenerateId() (string, error) {
	id, err := g.flake.NextID()
	if err != nil {
		return "", err
	}

	shortId := id >> 4 // reduce bits to 59, so base62 always returns <= 10 symbols string
	base62 := encodeBase62(shortId)

	builder := strings.Builder{}
	builder.Grow(10)
	builder.WriteString(base62)
	for builder.Len() < 10 {
		builder.WriteString("0")
	}

	return builder.String(), nil
}

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encodeBase62(n int64) string {
	if n == 0 {
		return string(alphabet[0])
	}

	var result []byte
	for n > 0 {
		result = append(result, alphabet[n%62])
		n /= 62
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
