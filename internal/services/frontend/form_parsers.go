package frontend

import (
	"fmt"
	"net/url"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
)

func anyToString(x interface{}) string {
	return fmt.Sprintf("%v", x)
}

func (s *service) stringToPointerToString(form url.Values, key string) *string {
	raw := form.Get(key)

	return &raw
}

func (s *service) stringToBool(form url.Values, key string) bool {
	raw := form.Get(key)

	x, err := strconv.ParseBool(raw)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	return x
}

func (s *service) stringToPointerToBool(form url.Values, key string) *bool {
	raw := form.Get(key)

	x, err := strconv.ParseBool(raw)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	return &x
}

func (s *service) stringToInt(form url.Values, key string) int {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int(i)

	return x
}

func (s *service) stringToPointerToInt(form url.Values, key string) *int {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int(i)

	return &x
}

func (s *service) stringToInt8(form url.Values, key string) int8 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int8(i)

	return x
}

func (s *service) stringToPointerToInt8(form url.Values, key string) *int8 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int8(i)

	return &x
}

func (s *service) stringToInt16(form url.Values, key string) int16 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int16(i)

	return x
}

func (s *service) stringToPointerToInt16(form url.Values, key string) *int16 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int16(i)

	return &x
}

func (s *service) stringToInt32(form url.Values, key string) int32 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int32(i)

	return x
}

func (s *service) stringToPointerToInt32(form url.Values, key string) *int32 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := int32(i)

	return &x
}

func (s *service) stringToInt64(form url.Values, key string) int64 {
	raw := form.Get(key)

	x, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	return x
}

func (s *service) stringToPointerToInt64(form url.Values, key string) *int64 {
	raw := form.Get(key)

	x, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	return &x
}

func (s *service) stringToUint(form url.Values, key string) uint {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint(i)

	return x
}

func (s *service) stringToPointerToUint(form url.Values, key string) *uint {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint(i)

	return &x
}

func (s *service) stringToUint8(form url.Values, key string) uint8 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint8(i)

	return x
}

func (s *service) stringToPointerToUint8(form url.Values, key string) *uint8 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint8(i)

	return &x
}

func (s *service) stringToUint16(form url.Values, key string) uint16 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint16(i)

	return x
}

func (s *service) stringToPointerToUint16(form url.Values, key string) *uint16 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint16(i)

	return &x
}

func (s *service) stringToUint32(form url.Values, key string) uint32 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint32(i)

	return x
}

func (s *service) stringToPointerToUint32(form url.Values, key string) *uint32 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint32(i)

	return &x
}

func (s *service) stringToUint64(form url.Values, key string) uint64 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint64(i)

	return x
}

func (s *service) stringToPointerToUint64(form url.Values, key string) *uint64 {
	raw := form.Get(key)

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := uint64(i)

	return &x
}

func (s *service) stringToFloat32(form url.Values, key string) float32 {
	raw := form.Get(key)

	i, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := float32(i)

	return x
}

func (s *service) stringToPointerToFloat32(form url.Values, key string) *float32 {
	raw := form.Get(key)

	i, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	x := float32(i)

	return &x
}

func (s *service) stringToFloat64(form url.Values, key string) float64 {
	raw := form.Get(key)

	x, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	return x
}

func (s *service) stringToPointerToFloat64(form url.Values, key string) *float64 {
	raw := form.Get(key)

	x, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, nil, "extracting form value")
	}

	return &x
}
